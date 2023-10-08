// Package core provides the core functions of YoMo.
package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/quic-go/quic-go"
	"github.com/yomorun/yomo/core/frame"
	"github.com/yomorun/yomo/pkg/frame-codec/y3codec"
	"github.com/yomorun/yomo/pkg/id"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

// Client is the abstraction of a YoMo-Client. a YoMo-Client can be
// Source, Upstream Zipper or StreamFunction.
type Client struct {
	name           string                     // name of the client
	clientID       string                     // id of the client
	clientType     ClientType                 // type of the client
	processor      func(*frame.DataFrame)     // function to invoke when data arrived
	receiver       func(*frame.BackflowFrame) // function to invoke when data is processed
	errorfn        func(error)                // function to invoke when error occured
	opts           *clientOptions
	logger         *slog.Logger
	tracerProvider oteltrace.TracerProvider

	// ctx and ctxCancel manage the lifecycle of client.
	ctx       context.Context
	ctxCancel context.CancelCauseFunc

	writeFrameChan chan frame.Frame

	// quic connection
	conn quic.Connection
	// frame stream
	fs *FrameStream
}

// NewClient creates a new YoMo-Client.
func NewClient(appName string, clientType ClientType, opts ...ClientOption) *Client {
	option := defaultClientOption()

	for _, o := range opts {
		o(option)
	}
	clientID := id.New()

	logger := option.logger.With("component", clientType.String(), "client_id", clientID, "client_name", appName)

	if option.credential != nil {
		logger.Info("use credential", "credential_name", option.credential.Name())
	}

	ctx, ctxCancel := context.WithCancelCause(context.Background())

	return &Client{
		name:           appName,
		clientID:       clientID,
		processor:      func(df *frame.DataFrame) { logger.Warn("the processor has not been set") },
		receiver:       func(bf *frame.BackflowFrame) { logger.Warn("the receiver has not been set") },
		clientType:     clientType,
		opts:           option,
		logger:         logger,
		tracerProvider: option.tracerProvider,
		errorfn:        func(err error) { logger.Error("client err", "err", err) },
		writeFrameChan: make(chan frame.Frame),
		ctx:            ctx,
		ctxCancel:      ctxCancel,
	}
}

type connectResult struct {
	conn quic.Connection
	fs   *FrameStream
	err  error
}

func newConnectResult(conn quic.Connection, fs *FrameStream, err error) *connectResult {
	return &connectResult{
		conn: conn,
		fs:   fs,
		err:  err,
	}
}

func (c *Client) connect(ctx context.Context, addr string) *connectResult {
	conn, err := quic.DialAddr(ctx, addr, c.opts.tlsConfig, c.opts.quicConfig)
	if err != nil {
		return newConnectResult(conn, nil, err)
	}

	stream, err := conn.OpenStream()
	if err != nil {
		return newConnectResult(conn, nil, err)
	}

	fs := NewFrameStream(stream, y3codec.Codec(), y3codec.PacketReadWriter())

	hf := &frame.HandshakeFrame{
		Name:            c.name,
		ID:              c.clientID,
		ClientType:      byte(c.clientType),
		ObserveDataTags: c.opts.observeDataTags,
		AuthName:        c.opts.credential.Name(),
		AuthPayload:     c.opts.credential.Payload(),
	}

	if err := fs.WriteFrame(hf); err != nil {
		return newConnectResult(conn, nil, err)
	}

	received, err := fs.ReadFrame()
	if err != nil {
		return newConnectResult(conn, nil, err)
	}

	switch received.Type() {
	case frame.TypeRejectedFrame:
		se := ErrAuthenticateFailed{received.(*frame.RejectedFrame).Message}
		return newConnectResult(conn, fs, se)
	case frame.TypeHandshakeAckFrame:
		return newConnectResult(conn, fs, nil)
	default:
		se := ErrAuthenticateFailed{
			fmt.Sprintf("authentication failed: read unexcepted frame, frame read: %s", received.Type().String()),
		}
		return newConnectResult(conn, fs, se)
	}
}

func (c *Client) runBackground(ctx context.Context, addr string, conn quic.Connection, fs *FrameStream) {
	reconnection := make(chan struct{})

	go c.handleReadFrames(fs, reconnection)

	for {
		select {
		case <-c.ctx.Done():
			fs.Close()
			return
		case <-ctx.Done():
			fs.Close()
			return
		case f := <-c.writeFrameChan:
			if err := fs.WriteFrame(f); err != nil {
				c.handleFrameError(err, reconnection)
			}
		case <-reconnection:
		reconnect:
			cr := c.connect(ctx, addr)
			if err := cr.err; err != nil {
				if errors.As(err, new(ErrAuthenticateFailed)) {
					return
				}
				c.logger.Error("reconnect to zipper error", "err", cr.err)
				time.Sleep(time.Second)
				goto reconnect
			}
			fs = cr.fs
			c.setConnection(&cr.conn)
			go c.handleReadFrames(fs, reconnection)
		}
	}
}

// Connect connect client to server.
func (c *Client) Connect(ctx context.Context, addr string) error {
	if c.clientType == ClientTypeStreamFunction && len(c.opts.observeDataTags) == 0 {
		return errors.New("yomo: streamFunction cannot observe data because the required tag has not been set")
	}

	c.logger = c.logger.With("zipper_addr", addr)

connect:
	result := c.connect(ctx, addr)
	if result.err != nil {
		if c.opts.connectUntilSucceed {
			c.logger.Error("failed to connect to zipper, trying to reconnect", "err", result.err)
			time.Sleep(time.Second)
			goto connect
		}
		c.logger.Error("can not connect to zipper", "err", result.err)
		return result.err
	}
	c.logger = c.logger.With("local_addr", result.conn.LocalAddr().String())
	c.logger.Info("connected to zipper")

	c.setConnection(&result.conn)
	c.setFrameStream(result.fs)

	go c.runBackground(ctx, addr, result.conn, result.fs)

	return nil
}

// WriteFrame write frame to client.
func (c *Client) WriteFrame(f frame.Frame) error {
	if c.opts.nonBlockWrite {
		return c.nonBlockWriteFrame(f)
	}
	return c.blockWriteFrame(f)
}

// blockWriteFrame writes frames in block mode, guaranteeing that frames are not lost.
func (c *Client) blockWriteFrame(f frame.Frame) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.writeFrameChan <- f:
	}
	return nil
}

// nonBlockWriteFrame writes frames in non-blocking mode, without guaranteeing that frames will not be lost.
func (c *Client) nonBlockWriteFrame(f frame.Frame) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.writeFrameChan <- f:
		return nil
	default:
		err := errors.New("yomo: client has lost connection")
		c.logger.Debug("failed to write frame", "frame_type", f.Type().String(), "error", err)
		return err
	}
}

// Close close the client.
func (c *Client) Close() error {
	// break runBackgroud() for-loop.
	c.ctxCancel(fmt.Errorf("%s: local shutdown", c.clientType.String()))

	return nil
}

// handleFrameError handles errors that occur during frame reading and writing by performing the following actions:
// Sending the error to the error function (errorfn).
// Closing the client if the connecion has been closed.
// Always attempting to reconnect if an error is encountered.
func (c *Client) handleFrameError(err error, reconnection chan<- struct{}) {
	if err == nil {
		return
	}

	c.errorfn(err)

	// exit client program if stream has be closed.
	if err == io.EOF {
		c.ctxCancel(fmt.Errorf("%s: remote shutdown", c.clientType.String()))
		return
	}

	// always attempting to reconnect if an error is encountered,
	// the error is mostly network error.
	select {
	case reconnection <- struct{}{}:
	default:
	}
}

// Wait waits client returning.
func (c *Client) Wait() {
	<-c.ctx.Done()
}

func (c *Client) handleReadFrames(fs *FrameStream, reconnection chan struct{}) {
	for {
		f, err := fs.ReadFrame()
		if err != nil {
			c.handleFrameError(err, reconnection)
			return
		}
		func() {
			defer func() {
				if e := recover(); e != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]

					perr := fmt.Errorf("%v", e)
					c.logger.Error("stream panic", "err", perr)
					c.errorfn(fmt.Errorf("yomo: stream panic: %v\n%s", perr, buf))
				}
			}()
			c.handleFrame(f)
		}()
	}
}

func (c *Client) handleFrame(f frame.Frame) {
	switch ff := f.(type) {
	case *frame.RejectedFrame:
		c.logger.Error("rejected error", "err", ff.Message)
		_ = c.Close()
	case *frame.DataFrame:
		c.processor(ff)
	case *frame.BackflowFrame:
		c.receiver(ff)
	case *frame.StreamFrame:
		// TODO: handle stream frame
		c.logger.Debug("receive stream frame", "stream_id", ff.StreamID, "conn_id", ff.ClientID, "tag", ff.Tag)
	default:
		c.logger.Error("received unexpected frame", "frame_type", f.Type().String())
	}
}

// SetDataFrameObserver sets the data frame handler.
func (c *Client) SetDataFrameObserver(fn func(*frame.DataFrame)) {
	c.processor = fn
}

// SetBackflowFrameObserver sets the backflow frame handler.
func (c *Client) SetBackflowFrameObserver(fn func(*frame.BackflowFrame)) {
	c.receiver = fn
}

// SetObserveDataTags set the data tag list that will be observed.
func (c *Client) SetObserveDataTags(tag ...frame.Tag) {
	c.opts.observeDataTags = tag
}

// Logger get client's logger instance, you can customize this using `yomo.WithLogger`
func (c *Client) Logger() *slog.Logger {
	return c.logger
}

// SetErrorHandler set error handler
func (c *Client) SetErrorHandler(fn func(err error)) {
	c.errorfn = fn
	c.logger.Debug("the error handler has been set")
}

// ClientID returns the ID of client.
func (c *Client) ClientID() string { return c.clientID }

// Name returns the name of client.
func (c *Client) Name() string { return c.name }

// FrameWriterConnection represents a frame writer that can connect to an addr.
type FrameWriterConnection interface {
	frame.Writer
	ClientID() string
	Name() string
	Close() error
	Connect(context.Context, string) error
}

// TracerProvider returns the tracer provider of client.
func (c *Client) TracerProvider() oteltrace.TracerProvider {
	if c.tracerProvider == nil {
		return nil
	}
	if reflect.ValueOf(c.tracerProvider).IsNil() {
		return nil
	}
	return c.tracerProvider
}

// ErrAuthenticateFailed be returned when client control stream authenticate failed.
type ErrAuthenticateFailed struct {
	ReasonFromServer string
}

// Error returns a string that represents the ErrAuthenticateFailed error for the implementation of the error interface.
func (e ErrAuthenticateFailed) Error() string {
	return e.ReasonFromServer
}

/*
// RequestStream request a stream from server.
func (c *Client) RequestStream() (quic.Stream, error) {
	// request data stream
	c.logger.Debug("client request data stream")
	dataStream, err := c.Connection().OpenStream()
	if err != nil {
		if err == io.EOF {
			c.logger.Info("client request data stream EOF")
			dataStream.Close()
			return nil, err
		}
		c.logger.Error("client request data stream error", "err", err)
		c.errorfn(err)
		return nil, err
	}
	c.logger.Debug("client write stream frame success", "stream_id", dataStream.StreamID())
	return dataStream, nil
}
*/

// PipeStream pipe a stream to server.
func (c *Client) PipeStream(ctx context.Context, dataStreamID string, stream io.Reader) error {
	c.logger.Debug("client pipe stream -- start")
	// for {
	qconn := c.Connection()
STREAM:
	dataStream, err := qconn.AcceptStream(ctx)
	if err != nil {
		c.logger.Error("client accept data stream error", "err", err)
		// c.errorfn(err)
		return err
	}
	// close data stream
	defer dataStream.Close()
	c.logger.Debug("client accept stream success", "stream_id", dataStream.StreamID())
	// read stream frame
	fs := NewFrameStream(dataStream, y3codec.Codec(), y3codec.PacketReadWriter())
	f, err := fs.ReadFrame()
	if err != nil {
		c.logger.Warn("failed to read data stream", "err", err)
		return err
	}
	c.logger.Debug("client read stream frame success", "stream_id", dataStream.StreamID())
	switch f.Type() {
	case frame.TypeStreamFrame:
		streamFrame := f.(*frame.StreamFrame)
		// if stream id is same, pipe stream
		if streamFrame.ID != dataStreamID {
			c.logger.Debug(
				"stream id is not same, continue",
				"stream_id", dataStream.StreamID(),
				"datastream_id", dataStreamID,
				"received_id", streamFrame.ID,
				"client_id", streamFrame.ClientID,
				"tag", streamFrame.Tag,
			)
			goto STREAM
		}
		c.logger.Info(
			"!!!pipe stream is ready!!!",
			"remote_addr", qconn.RemoteAddr().String(),
			"datastream_id", streamFrame.ID,
			"stream_id", dataStream.StreamID(),
			"client_id", streamFrame.ClientID,
			"id", streamFrame.ID,
			"tag", streamFrame.Tag,
		)
		// pipe stream
		n, err := io.Copy(dataStream, stream)
		if err != nil {
			c.logger.Error("!!!pipe stream error!!!", "err", err)
			return err
		}
		c.logger.Info("!!!pipe stream success!!!",
			"remote_addr", qconn.RemoteAddr().String(),
			"id", streamFrame.ID,
			"stream_id", dataStream.StreamID(),
			"client_id", streamFrame.ClientID,
			"tag", streamFrame.Tag,
			"n", n,
		)
	default:
		c.logger.Error("!!!unexpected frame!!!", "unexpected_frame_type", f.Type().String())
		return errors.New("unexpected frame")
	}
	// }
	c.logger.Debug("client pipe stream -- end")
	return nil
}

// Connection returns the connection of client.
func (c *Client) Connection() quic.Connection {
	conn := (*quic.Connection)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.conn))))
	if conn != nil {
		return *conn
	}
	return nil
}

// setConnection set the connection of client.
func (c *Client) setConnection(conn *quic.Connection) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.conn)), unsafe.Pointer(conn))
}

// FrameStream returns the FrameStream of client.
func (c *Client) FrameStream() *FrameStream {
	return (*FrameStream)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.fs))))
}

// setFrameStream set the FrameStream of client.
func (c *Client) setFrameStream(fs *FrameStream) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.fs)), unsafe.Pointer(fs))
}
