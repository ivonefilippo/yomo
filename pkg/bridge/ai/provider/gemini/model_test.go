package gemini

import (
	"reflect"
	"testing"

	"github.com/yomorun/yomo/ai"
)

func TestConvertPropertyToStandard(t *testing.T) {
	properties := map[string]*Property{
		"prop1": {Type: "type1", Description: "desc1"},
		"prop2": {Type: "type2", Description: "desc2"},
	}

	expected := map[string]*ai.ParameterProperty{
		"prop1": {Type: "type1", Description: "desc1"},
		"prop2": {Type: "type2", Description: "desc2"},
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
			"prop1": {Type: "type1", Description: "desc1"},
			"prop2": {Type: "type2", Description: "desc2"},
		},
		Required: []string{"prop1"},
	}

	expected := &ai.FunctionParameters{
		Type: "type1",
		Properties: map[string]*ai.ParameterProperty{
			"prop1": {Type: "type1", Description: "desc1"},
			"prop2": {Type: "type2", Description: "desc2"},
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
				"prop1": {Type: "type1", Description: "desc1"},
				"prop2": {Type: "type2", Description: "desc2"},
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
				"prop1": {Type: "type1", Description: "desc1"},
				"prop2": {Type: "type2", Description: "desc2"},
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
		"prop1": {Type: "type1", Description: "desc1"},
		"prop2": {Type: "type2", Description: "desc2"},
	}

	expected := map[string]*Property{
		"prop1": {Type: "type1", Description: "desc1"},
		"prop2": {Type: "type2", Description: "desc2"},
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
			"prop1": {Type: "type1", Description: "desc1"},
			"prop2": {Type: "type2", Description: "desc2"},
		},
		Required: []string{"prop1"},
	}

	expected := &FunctionParameters{
		Type: "type1",
		Properties: map[string]*Property{
			"prop1": {Type: "type1", Description: "desc1"},
			"prop2": {Type: "type2", Description: "desc2"},
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
				"prop1": {Type: "type1", Description: "desc1"},
				"prop2": {Type: "type2", Description: "desc2"},
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
				"prop1": {Type: "type1", Description: "desc1"},
				"prop2": {Type: "type2", Description: "desc2"},
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

func TestGenerateJSONSchemaArguments(t *testing.T) {
	args := Args{
		"arg1": "value1",
		"arg2": "value2",
	}

	expected := `{"arg1":"value1","arg2":"value2"}`

	result := generateJSONSchemaArguments(args)

	if result != expected {
		t.Errorf("generateJSONSchemaArguments() = %v, want %v", result, expected)
	}
}

func TestGenerateJSONSchemaArguments_EmptyArgs(t *testing.T) {
	args := Args{}

	expected := `{}`

	result := generateJSONSchemaArguments(args)

	if result != expected {
		t.Errorf("generateJSONSchemaArguments() = %v, want %v", result, expected)
	}
}

func TestParseAPIResponseBody(t *testing.T) {
	respBody := []byte(`[{
		"candidates": [
			{
				"content": {
					"parts": [
						{
							"functionCall": {
								"name": "find_theaters",
								"args": {
									"movie": "Barbie",
									"location": "Mountain View, CA"
								}
							}
						}
					]
				},
				"finishReason": "STOP",
				"safetyRatings": [
					{
						"category": "HARM_CATEGORY_HARASSMENT",
						"probability": "NEGLIGIBLE"
					},
					{
						"category": "HARM_CATEGORY_HATE_SPEECH",
						"probability": "NEGLIGIBLE"
					},
					{
						"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
						"probability": "NEGLIGIBLE"
					},
					{
						"category": "HARM_CATEGORY_DANGEROUS_CONTENT",
						"probability": "NEGLIGIBLE"
					}
				]
			}
		],
		"usageMetadata": {
			"promptTokenCount": 9,
			"totalTokenCount": 9
		}
	}]`)

	expected := []*Response{
		{
			Candidates: []Candidate{
				{
					Content: CandidateContent{
						Parts: []Part{
							{
								FunctionCall: FunctionCall{
									Name: "find_theaters",
									Args: map[string]interface{}{
										"movie":    "Barbie",
										"location": "Mountain View, CA",
									},
								},
							},
						},
					},
					FinishReason: "STOP",
					SafetyRatings: []CandidateSafetyRating{
						{Category: "HARM_CATEGORY_HARASSMENT", Probability: "NEGLIGIBLE"},
						{Category: "HARM_CATEGORY_HATE_SPEECH", Probability: "NEGLIGIBLE"},
						{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Probability: "NEGLIGIBLE"},
						{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Probability: "NEGLIGIBLE"},
					},
				},
			},
			UsageMetadata: UsageMetadata{
				PromptTokenCount: 9,
				TotalTokenCount:  9,
			},
		},
	}

	result, err := parseAPIResponseBody(respBody)
	if err != nil {
		t.Fatalf("parseAPIResponseBody() error = %v, wantErr %v", err, false)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("parseAPIResponseBody() = %v, want %v", result, expected)
	}
}

func TestParseAPIResponseBody_InvalidJSON(t *testing.T) {
	respBody := []byte(`invalid json`)

	_, err := parseAPIResponseBody(respBody)
	if err == nil {
		t.Errorf("parseAPIResponseBody() error = %v, wantErr %v", err, true)
	}
}

func TestParseToolCallFromResponse(t *testing.T) {
	resp := []*Response{
		{
			Candidates: []Candidate{
				{
					Content: CandidateContent{
						Parts: []Part{
							{
								FunctionCall: FunctionCall{
									Name: "find_theaters",
									Args: map[string]interface{}{
										"location": "Mountain View, CA",
										"movie":    "Barbie",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expected := []*ai.ToolCall{
		{
			Function: &ai.FunctionDefinition{
				Name:      "find_theaters",
				Arguments: "{\"location\":\"Mountain View, CA\",\"movie\":\"Barbie\"}",
			},
			ID:   "cc-gemini-id",
			Type: "cc-function",
		},
	}

	result := parseToolCallFromResponse(resp)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("parseToolCallFromResponse() = %v, want %v", result, expected)
	}
}
