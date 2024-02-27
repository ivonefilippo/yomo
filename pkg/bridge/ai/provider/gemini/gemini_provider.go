package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/yomorun/yomo/ai"
	"github.com/yomorun/yomo/core/ylog"
	baseProvider "github.com/yomorun/yomo/pkg/bridge/ai"
)

var (
	fns sync.Map
)

type connectedFn struct {
	connID uint64
	tag    uint32
	fd     *FunctionDeclaration
}

func init() {
	fns = sync.Map{}
}

// GeminiProvider is the provider for Gemini
type GeminiProvider struct {
	APIKey string
}

var _ = baseProvider.LLMProvider(&GeminiProvider{})

// Name returns the name of the provider
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// GetChatCompletions get chat completions for ai service
func (p *GeminiProvider) GetChatCompletions(userInstruction string) (*ai.InvokeResponse, error) {
	// check if there are any tool calls attached, if no, return error
	isEmpty := true
	fns.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})

	if isEmpty {
		ylog.Error("-----tools is empty")
		return &ai.InvokeResponse{Content: "no toolCalls"}, ai.ErrNoFunctionCall
	}

	// prepare request body
	body := &RequestBody{}

	// prepare contents
	body.Contents.Role = "user"
	body.Contents.Parts.Text = userInstruction

	// prepare tools
	toolCalls := make([]*FunctionDeclaration, 0)
	fns.Range(func(_, value interface{}) bool {
		fn := value.(*connectedFn)
		toolCalls = append(toolCalls, fn.fd)
		return true
	})
	body.Tools = append(make([]Tool, 0), Tool{
		FunctionDeclarations: toolCalls,
	})

	// request API
	jsonBody, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error preparing request body:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", p.getApiUrl(), bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Error creating new request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gemini provider api response status code is %d", resp.StatusCode)
	}

	// parse response
	var response []Response
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		fmt.Println("Error parsing response body:", err)
		return nil, err
	}

	// Now you can access the data in the response
	calls := make([]ai.ToolCall, 0)
	result := &ai.InvokeResponse{}
	if len(calls) == 0 {
		return result, ai.ErrNoFunctionCall
	}
	for _, candidate := range response[0].Candidates {
		fn := candidate.Content.Parts[0].FunctionCall
		call := ai.ToolCall{
			ID:   "cc-gemini-id",
			Type: "cc-function",
			Function: &ai.FunctionDefinition{
				Name:      fn.Name,
				Arguments: p.generateJSONSchemaArguments(fn.Args),
			},
		}
		fmt.Printf("Function name: %s", fn.Name)
		fmt.Printf("Function args: %+v", fn.Args)
		calls = append(calls, call)
	}

	// messages
	return result, nil
}

// RegisterFunction registers the llm function
func (p *GeminiProvider) RegisterFunction(tag uint32, functionDefinition *ai.FunctionDefinition, connID uint64) error {
	fns.Store(connID, &connectedFn{
		connID: connID,
		tag:    tag,
		fd:     convertStandardToFunctionDeclaration(functionDefinition),
	})

	return nil
}

// UnregisterFunction unregister the llm function
func (p *GeminiProvider) UnregisterFunction(name string, connID uint64) error {
	fns.Delete(connID)
	return nil
}

// ListToolCalls lists the llm tool calls
func (p *GeminiProvider) ListToolCalls() (map[uint32]ai.ToolCall, error) {
	result := make(map[uint32]ai.ToolCall)

	tmp := make(map[uint32]*FunctionDeclaration)
	fns.Range(func(_, value any) bool {
		fn := value.(*connectedFn)
		tmp[fn.tag] = fn.fd
		return true
	})

	return result, nil
}

// GetOverview returns the overview of the AI functions, key is the tag, value is the function definition
func (p *GeminiProvider) GetOverview() (*ai.OverviewResponse, error) {
	isEmpty := true
	fns.Range(func(_, _ any) bool {
		isEmpty = false
		return false
	})

	result := &ai.OverviewResponse{
		Functions: make(map[uint32]*ai.FunctionDefinition),
	}

	if isEmpty {
		return result, nil
	}

	fns.Range(func(_, value any) bool {
		fn := value.(*connectedFn)
		result.Functions[fn.tag] = convertFunctionDeclarationToStandard(fn.fd)
		return true
	})

	return result, nil
}

// getApiUrl returns the gemini generateContent API url
func (p *GeminiProvider) getApiUrl() string {
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", p.APIKey)
}

// generateJSONSchemaArguments generates the JSON schema arguments from OpenAPI compatible arguments
// https://ai.google.dev/docs/function_calling#how_it_works
func (p *GeminiProvider) generateJSONSchemaArguments(args Args) string {
	schema := make(map[string]interface{})

	for k, v := range args {
		schema[k] = v
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return ""
	}

	return string(schemaJSON)
}

// NewGeminiProvider creates a new GeminiProvider
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		APIKey: apiKey,
	}
}

// New creates a new GeminiProvider
func New() *GeminiProvider {
	return &GeminiProvider{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	}
}
