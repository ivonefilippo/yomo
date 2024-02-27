package gemini

type Response struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

// Candidate is the element of Response
type Candidate struct {
	Content       CandidateContent        `json:"content"`
	FinishReason  string                  `json:"finishReason"`
	SafetyRatings []CandidateSafetyRating `json:"safetyRatings"`
}

// CandidateContent is the content of Candidate
type CandidateContent struct {
	Parts []Part `json:"parts"`
}

// CandidateSafetyRating is the safetyRatings of Candidate
type CandidateSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// Part is the element of CandidateContent
type Part struct {
	FunctionCall FunctionCall `json:"functionCall"`
}

// FunctionCall is the functionCall of Part
type FunctionCall struct {
	Name string `json:"name"`
	Args Args   `json:"args"`
}

// Args is the args of FunctionCall
type Args map[string]interface{}

// UsageMetadata is the token usage in Response
type UsageMetadata struct {
	PromptTokenCount int `json:"promptTokenCount"`
	TotalTokenCount  int `json:"totalTokenCount"`
}
