package gemini

type RequestBody struct {
	Contents struct {
		Role  string `json:"role"`
		Parts Parts  `json:"parts"`
	} `json:"contents"`
	Tools []Tool `json:"tools"`
}

// Parts is the contents.parts in RequestBody
type Parts struct {
	Text string `json:"text"`
}

// Tool is the element of tools in RequestBody
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"function_declarations"`
}

// FunctionDeclaration is the element of Tool
type FunctionDeclaration struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  FunctionParameters `json:"parameters"`
}

// FunctionParameters is the parameters of FunctionDeclaration
type FunctionParameters struct {
	Type       string              `json:"type"`
	Properties ParameterProperties `json:"properties"`
	Required   []string            `json:"required"`
}

// ParameterProperties is the properties of FunctionParameters
type ParameterProperties struct {
	Location    Property `json:"location"`
	Description Property `json:"description"`
	Movie       Property `json:"movie"`
	Theater     Property `json:"theater"`
	Date        Property `json:"date"`
}

// Property is the element of ParameterProperties
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
