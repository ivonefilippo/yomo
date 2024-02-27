package gemini

import (
	"reflect"
	"testing"

	"github.com/yomorun/yomo/ai"
)

func TestConvertPropertyToStandard(t *testing.T) {
	properties := map[string]*Property{
		"prop1": &Property{Type: "type1", Description: "desc1"},
		"prop2": &Property{Type: "type2", Description: "desc2"},
	}

	expected := map[string]*ai.ParameterProperty{
		"prop1": &ai.ParameterProperty{Type: "type1", Description: "desc1"},
		"prop2": &ai.ParameterProperty{Type: "type2", Description: "desc2"},
	}

	result := convertPropertyToStandard(properties)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertPropertyToStandard() = %v, want %v", result, expected)
	}
}

func TestConvertPropertyToStandard_NilInput(t *testing.T) {
	result := convertPropertyToStandard(nil)

	if result != nil {
		t.Errorf("convertPropertyToStandard() = %v, want %v", result, nil)
	}
}

func TestConvertFunctionParametersToStandard(t *testing.T) {
	parameters := &FunctionParameters{
		Type: "type1",
		Properties: map[string]*Property{
			"prop1": &Property{Type: "type1", Description: "desc1"},
			"prop2": &Property{Type: "type2", Description: "desc2"},
		},
		Required: []string{"prop1"},
	}

	expected := &ai.FunctionParameters{
		Type: "type1",
		Properties: map[string]*ai.ParameterProperty{
			"prop1": &ai.ParameterProperty{Type: "type1", Description: "desc1"},
			"prop2": &ai.ParameterProperty{Type: "type2", Description: "desc2"},
		},
		Required: []string{"prop1"},
	}

	result := convertFunctionParametersToStandard(parameters)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertFunctionParametersToStandard() = %v, want %v", result, expected)
	}
}

func TestConvertFunctionParametersToStandard_NilInput(t *testing.T) {
	result := convertFunctionParametersToStandard(nil)

	if result != nil {
		t.Errorf("convertFunctionParametersToStandard() = %v, want %v", result, nil)
	}
}

func TestConvertFunctionDeclarationToStandard(t *testing.T) {
	functionDeclaration := &FunctionDeclaration{
		Name:        "function1",
		Description: "desc1",
		Parameters: &FunctionParameters{
			Type: "type1",
			Properties: map[string]*Property{
				"prop1": &Property{Type: "type1", Description: "desc1"},
				"prop2": &Property{Type: "type2", Description: "desc2"},
			},
			Required: []string{"prop1"},
		},
	}

	expected := &ai.FunctionDefinition{
		Name:        "function1",
		Description: "desc1",
		Parameters: &ai.FunctionParameters{
			Type: "type1",
			Properties: map[string]*ai.ParameterProperty{
				"prop1": &ai.ParameterProperty{Type: "type1", Description: "desc1"},
				"prop2": &ai.ParameterProperty{Type: "type2", Description: "desc2"},
			},
			Required: []string{"prop1"},
		},
	}

	result := convertFunctionDeclarationToStandard(functionDeclaration)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertFunctionDeclarationToStandard() = %v, want %v", result, expected)
	}
}

func TestConvertFunctionDeclarationToStandard_NilInput(t *testing.T) {
	result := convertFunctionDeclarationToStandard(nil)

	if result != nil {
		t.Errorf("convertFunctionDeclarationToStandard() = %v, want %v", result, nil)
	}
}

func TestConvertStandardToProperty(t *testing.T) {
	properties := map[string]*ai.ParameterProperty{
		"prop1": &ai.ParameterProperty{Type: "type1", Description: "desc1"},
		"prop2": &ai.ParameterProperty{Type: "type2", Description: "desc2"},
	}

	expected := map[string]*Property{
		"prop1": &Property{Type: "type1", Description: "desc1"},
		"prop2": &Property{Type: "type2", Description: "desc2"},
	}

	result := convertStandardToProperty(properties)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertStandardToProperty() = %v, want %v", result, expected)
	}
}

func TestConvertStandardToProperty_NilInput(t *testing.T) {
	result := convertStandardToProperty(nil)

	if result != nil {
		t.Errorf("convertStandardToProperty() = %v, want %v", result, nil)
	}
}

func TestConvertStandardToFunctionParameters(t *testing.T) {
	parameters := &ai.FunctionParameters{
		Type: "type1",
		Properties: map[string]*ai.ParameterProperty{
			"prop1": &ai.ParameterProperty{Type: "type1", Description: "desc1"},
			"prop2": &ai.ParameterProperty{Type: "type2", Description: "desc2"},
		},
		Required: []string{"prop1"},
	}

	expected := &FunctionParameters{
		Type: "type1",
		Properties: map[string]*Property{
			"prop1": &Property{Type: "type1", Description: "desc1"},
			"prop2": &Property{Type: "type2", Description: "desc2"},
		},
		Required: []string{"prop1"},
	}

	result := convertStandardToFunctionParameters(parameters)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertStandardToFunctionParameters() = %v, want %v", result, expected)
	}
}

func TestConvertStandardToFunctionParameters_NilInput(t *testing.T) {
	result := convertStandardToFunctionParameters(nil)

	if result != nil {
		t.Errorf("convertStandardToFunctionParameters() = %v, want %v", result, nil)
	}
}

func TestConvertStandardToFunctionDeclaration(t *testing.T) {
	functionDefinition := &ai.FunctionDefinition{
		Name:        "function1",
		Description: "desc1",
		Parameters: &ai.FunctionParameters{
			Type: "type1",
			Properties: map[string]*ai.ParameterProperty{
				"prop1": &ai.ParameterProperty{Type: "type1", Description: "desc1"},
				"prop2": &ai.ParameterProperty{Type: "type2", Description: "desc2"},
			},
			Required: []string{"prop1"},
		},
	}

	expected := &FunctionDeclaration{
		Name:        "function1",
		Description: "desc1",
		Parameters: &FunctionParameters{
			Type: "type1",
			Properties: map[string]*Property{
				"prop1": &Property{Type: "type1", Description: "desc1"},
				"prop2": &Property{Type: "type2", Description: "desc2"},
			},
			Required: []string{"prop1"},
		},
	}

	result := convertStandardToFunctionDeclaration(functionDefinition)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertStandardToFunctionDeclaration() = %v, want %v", result, expected)
	}
}

func TestConvertStandardToFunctionDeclaration_NilInput(t *testing.T) {
	result := convertStandardToFunctionDeclaration(nil)

	if result != nil {
		t.Errorf("convertStandardToFunctionDeclaration() = %v, want %v", result, nil)
	}
}
