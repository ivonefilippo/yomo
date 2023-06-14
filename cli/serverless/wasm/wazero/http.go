package wazero

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	wasmhttp "github.com/yomorun/yomo/cli/serverless/wasm/http"
)

func ExportHTTPHostFuncs(builder wazero.HostModuleBuilder) {
	builder.
		// get
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(Get),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmhttp.WasmFuncHTTPGet)
}

func Get(ctx context.Context, m api.Module, stack []uint64) {
	pointer := uint32(stack[0])
	length := uint32(stack[1])
	buf, ok := m.Memory().Read(pointer, length)
	if !ok {
		log.Printf("Memory.Read(%d, %d) out of range\n", pointer, length)
		stack[0] = 1
		return
	}
	url := make([]byte, length)
	copy(url, buf)

	req, err := http.NewRequest("GET", string(url), nil)
	if err != nil {
		log.Printf("http.NewRequest(%s) error\n", url)
		stack[0] = 1
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("http.Get(%s) error\n", url)
		stack[0] = 2
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll error\n")
		stack[0] = 3
		return
	}
	log.Printf("http.Get(%s) success: %s\n", url, body)
	stack[0] = 0
}
