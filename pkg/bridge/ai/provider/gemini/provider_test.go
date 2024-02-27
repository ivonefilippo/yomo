package gemini

import (
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/yomorun/yomo/ai"
)

func TestGeminiProvider_Name(t *testing.T) {
	provider := &GeminiProvider{}

	name := provider.Name()

	if name != "gemini" {
		t.Errorf("Name() = %v, want %v", name, "gemini")
	}
}

func TestGeminiProvider_getApiUrl(t *testing.T) {
	provider := &GeminiProvider{
		APIKey: "test-api-key",
	}

	expected := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=test-api-key"

	result := provider.getApiUrl()

	if result != expected {
		t.Errorf("getApiUrl() = %v, want %v", result, expected)
	}
}

func TestNewGeminiProvider(t *testing.T) {
	apiKey := "test-api-key"
	provider := NewGeminiProvider(apiKey)

	if provider.APIKey != apiKey {
		t.Errorf("NewGeminiProvider() = %v, want %v", provider.APIKey, apiKey)
	}
}

func TestNew(t *testing.T) {
	// Set up
	expectedAPIKey := "test-api-key"
	os.Setenv("GEMINI_API_KEY", expectedAPIKey)

	// Call the function under test
	provider := New()

	// Check the result
	if provider.APIKey != expectedAPIKey {
		t.Errorf("New() = %v, want %v", provider.APIKey, expectedAPIKey)
	}
}

func TestNew_NoEnvVar(t *testing.T) {
	// Set up
	os.Unsetenv("GEMINI_API_KEY")

	// Call the function under test
	provider := New()

	// Check the result
	if provider.APIKey != "" {
		t.Errorf("New() = %v, want %v", provider.APIKey, "")
	}
}

func TestGeminiProvider_GetOverview_Empty(t *testing.T) {
	provider := &GeminiProvider{}

	result, err := provider.GetOverview()

	if err != nil {
		t.Errorf("GetOverview() error = %v, wantErr %v", err, nil)
		return
	}

	if len(result.Functions) != 0 {
		t.Errorf("GetOverview() = %v, want %v", len(result.Functions), 0)
	}
}

func TestGeminiProvider_GetOverview_NotEmpty(t *testing.T) {
	provider := &GeminiProvider{}

	// Add a function to the fns map
	fns.Store("test", &connectedFn{
		tag: 1,
		fd: &FunctionDeclaration{
			Name:        "function1",
			Description: "desc1",
			Parameters: &FunctionParameters{
				Type: "type1",
				Properties: map[string]*Property{
					"prop1": {Type: "type1", Description: "desc1"},
					"prop2": {Type: "type2", Description: "desc2"},
				},
				Required: []string{"prop1"},
			},
		},
	})

	result, err := provider.GetOverview()

	if err != nil {
		t.Errorf("GetOverview() error = %v, wantErr %v", err, nil)
		return
	}

	if len(result.Functions) != 1 {
		t.Errorf("GetOverview() = %v, want %v", len(result.Functions), 1)
	}
}

func TestGeminiProvider_ListToolCalls_Empty(t *testing.T) {
	fns = sync.Map{}
	provider := &GeminiProvider{}

	result, err := provider.ListToolCalls()

	if err != nil {
		t.Errorf("ListToolCalls() error = %v, wantErr %v", err, nil)
		return
	}

	if len(result) != 0 {
		t.Errorf("ListToolCalls() = %v, want %v", len(result), 0)
	}
}

func TestGeminiProvider_ListToolCalls_NotEmpty(t *testing.T) {
	provider := &GeminiProvider{}

	// Add a function to the fns map
	fns.Store("test", &connectedFn{
		tag: 1,
		fd: &FunctionDeclaration{
			Name:        "function1",
			Description: "desc1",
			Parameters: &FunctionParameters{
				Type: "type1",
				Properties: map[string]*Property{
					"prop1": {Type: "type1", Description: "desc1"},
					"prop2": {Type: "type2", Description: "desc2"},
				},
				Required: []string{"prop1"},
			},
		},
	})

	result, err := provider.ListToolCalls()

	if err != nil {
		t.Errorf("ListToolCalls() error = %v, wantErr %v", err, nil)
		return
	}

	if len(result) != 1 {
		t.Errorf("ListToolCalls() = %v, want %v", len(result), 1)
	}

	// TearDown
	fns = sync.Map{}
}

func TestGeminiProvider_RegisterFunction(t *testing.T) {
	provider := &GeminiProvider{}
	tag := uint32(1)
	connID := uint64(1)
	functionDefinition := &ai.FunctionDefinition{
		Name:        "function1",
		Description: "desc1",
		Parameters: &ai.FunctionParameters{
			Type: "type1",
			Properties: map[string]*ai.ParameterProperty{
				"prop1": {Type: "type1", Description: "desc1"},
				"prop2": {Type: "type2", Description: "desc2"},
			},
			Required: []string{"prop1"},
		},
	}

	err := provider.RegisterFunction(tag, functionDefinition, connID)

	if err != nil {
		t.Errorf("RegisterFunction() error = %v, wantErr %v", err, nil)
		return
	}

	value, ok := fns.Load(connID)
	if !ok {
		t.Errorf("RegisterFunction() did not store the function correctly")
		return
	}

	cf := value.(*connectedFn)
	if cf.connID != connID || cf.tag != tag || !reflect.DeepEqual(cf.fd, convertStandardToFunctionDeclaration(functionDefinition)) {
		t.Errorf("RegisterFunction() = %v, want %v", cf, &connectedFn{
			connID: connID,
			tag:    tag,
			fd:     convertStandardToFunctionDeclaration(functionDefinition),
		})
	}
}

func TestGeminiProvider_UnregisterFunction(t *testing.T) {
	provider := &GeminiProvider{}
	connID := uint64(1)

	// Add a function to the fns map
	fns.Store(connID, &connectedFn{
		tag: 1,
		fd: &FunctionDeclaration{
			Name:        "function1",
			Description: "desc1",
			Parameters: &FunctionParameters{
				Type: "type1",
				Properties: map[string]*Property{
					"prop1": {Type: "type1", Description: "desc1"},
					"prop2": {Type: "type2", Description: "desc2"},
				},
				Required: []string{"prop1"},
			},
		},
	})

	err := provider.UnregisterFunction("function1", connID)

	if err != nil {
		t.Errorf("UnregisterFunction() error = %v, wantErr %v", err, nil)
		return
	}

	_, ok := fns.Load(connID)
	if ok {
		t.Errorf("UnregisterFunction() did not remove the function correctly")
	}

	// TearDown
	fns = sync.Map{}
}
