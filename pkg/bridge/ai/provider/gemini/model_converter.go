package gemini

import "github.com/yomorun/yomo/ai"

func convertStandardToFunctionDeclaration(functionDefinition *ai.FunctionDefinition) *FunctionDeclaration {
	if functionDefinition == nil {
		return nil
	}

	return &FunctionDeclaration{
		Name:        functionDefinition.Name,
		Description: functionDefinition.Description,
		Parameters:  convertStandardToFunctionParameters(functionDefinition.Parameters),
	}
}

func convertFunctionDeclarationToStandard(functionDefinition *FunctionDeclaration) *ai.FunctionDefinition {
	if functionDefinition == nil {
		return nil
	}

	return &ai.FunctionDefinition{
		Name:        functionDefinition.Name,
		Description: functionDefinition.Description,
		Parameters:  convertFunctionParametersToStandard(functionDefinition.Parameters),
	}
}

func convertStandardToFunctionParameters(parameters *ai.FunctionParameters) FunctionParameters {
	if parameters == nil {
		return FunctionParameters{}
	}

	return FunctionParameters{
		Type:       parameters.Type,
		Properties: convertStandardToProperty(parameters.Properties),
		Required:   parameters.Required,
	}
}

func convertFunctionParametersToStandard(parameters FunctionParameters) *ai.FunctionParameters {
	return &ai.FunctionParameters{
		Type:       parameters.Type,
		Properties: convertPropertyToStandard(parameters.Properties),
		Required:   parameters.Required,
	}
}

func convertStandardToProperty(properties map[string]*ai.ParameterProperty) map[string]Property {
	if properties == nil {
		return nil
	}

	result := make(map[string]Property)
	for k, v := range properties {
		result[k] = Property{
			Type:        v.Type,
			Description: v.Description,
		}
	}
	return result
}

func convertPropertyToStandard(properties map[string]Property) map[string]*ai.ParameterProperty {
	if properties == nil {
		return nil
	}

	result := make(map[string]*ai.ParameterProperty)
	for k, v := range properties {
		result[k] = &ai.ParameterProperty{
			Type:        v.Type,
			Description: v.Description,
		}
	}
	return result
}
