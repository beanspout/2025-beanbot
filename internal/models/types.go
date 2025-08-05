package models

// TroubleshootingData represents the structure of our JSON data
type TroubleshootingData struct {
	ErrorCodes   []ErrorCode   `json:"error_codes"`
	CommonIssues []CommonIssue `json:"common_issues"`
}

// ErrorCode represents an error code in the knowledge database
type ErrorCode struct {
	Code                   string   `json:"code"`
	Description            string   `json:"description"`
	Category               string   `json:"category"`
	Severity               string   `json:"severity"`
	TroubleshootingSteps   []string `json:"troubleshooting_steps"`
	RelatedComponents      []string `json:"related_components"`
	DocumentationReference string   `json:"documentation_reference"`
}

// CommonIssue represents a common issue in the knowledge database
type CommonIssue struct {
	Issue     string   `json:"issue"`
	Symptoms  []string `json:"symptoms"`
	Solutions []string `json:"solutions"`
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
