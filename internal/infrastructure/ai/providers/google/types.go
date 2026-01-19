package google

// Generate response types
type generateResponse struct {
	Candidates    []candidate    `json:"candidates"`
	UsageMetadata *usageMetadata `json:"usageMetadata,omitempty"`
}

type candidate struct {
	Content      content `json:"content"`
	FinishReason string  `json:"finishReason"`
	Index        int     `json:"index"`
}

type content struct {
	Parts []part `json:"parts"`
	Role  string `json:"role"`
}

type part struct {
	Text         string        `json:"text,omitempty"`
	FunctionCall *functionCall `json:"functionCall,omitempty"`
}

type functionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// Embedding response types
type embeddingResponse struct {
	Embedding embedding `json:"embedding"`
}

type embedding struct {
	Values []float64 `json:"values"`
}

// Error response types
type errorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
