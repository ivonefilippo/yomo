package ai

import (
	"sync"

	"github.com/yomorun/yomo/ai"
)

// AIProvider provides an interface to the llm providers
type AIProvider interface {
	// Name returns the name of the llm provider
	Name() string
	// RegisterFunction registers the llm function
	RegisterFunction(tag uint32, functionDefinition *ai.FunctionDefinition, connID string) error
	// UnregisterFunction unregister the llm function
	UnregisterFunction(name, connID string) error
	// ListToolCalls lists the llm tool calls
	ListToolCalls() (map[uint32]ai.ToolCall, error)
}

var (
	providers       sync.Map
	defaultProvider AIProvider
)

// RegisterProvider registers the llm provider
func RegisterProvider(provider AIProvider) {
	if provider != nil {
		providers.Store(provider.Name(), provider)
	}
}

// ListProviders returns the list of llm providers
func ListProviders() []string {
	var names []string
	providers.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

// SetDefaultProvider sets the default llm provider
func SetDefaultProvider(name string) {
	provider := GetProvider(name)
	if provider != nil {
		defaultProvider = provider
	}
}

// GetProvider returns the llm provider by name
func GetProvider(name string) AIProvider {
	if provider, ok := providers.Load(name); ok {
		return provider.(AIProvider)
	}
	return nil
}

// GetDefaultProvider returns the default AI provider
func GetDefaultProvider() (AIProvider, error) {
	if defaultProvider != nil {
		return defaultProvider, nil
	}
	names := ListProviders()
	if len(names) > 0 {
		p := GetProvider(names[0])
		if p != nil {
			return p, nil
		}
	}
	return nil, ErrNotExistsProvider
}
