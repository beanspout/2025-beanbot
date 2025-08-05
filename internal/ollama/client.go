package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/NZ26RQ_gme/lsie-beanbot/internal/models"
)

// Client handles communication with Ollama
type Client struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewClient creates a new Ollama client
func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second, // 2 minute timeout for model response generation
		},
	}
}

// TestConnection tests the connection to Ollama
func (oc *Client) TestConnection() bool {
	resp, err := oc.client.Get(oc.baseURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// FindAvailableModel tries to find an available model
func (oc *Client) FindAvailableModel() (bool, string) {
	// Check models in order, starting with llama3.2:1b as default
	models := []string{
		"llama3.2:1b", // Default model
		"llama3.2:3b",
		"llama3.2",
		"llama3.2:latest",
		// Other popular models
		"gemma3:1b",
		"gemma2:2b",
		"gemma2:1b",
		"phi3:mini",
		"llama3.1:8b",
		"llama3.1:7b",
		"llama2:7b",
		"mistral:7b",
		"qwen2.5:1.5b",
		"codellama:7b",
	}

	for _, model := range models {
		if oc.testModel(model) {
			return true, model
		}
	}

	log.Printf("[DEBUG] No known models found. Consider installing: ollama pull llama3.2:1b")
	return false, ""
}

// SetModel sets the model to use
func (oc *Client) SetModel(model string) {
	oc.model = model
}

// GetCurrentModel returns the currently selected model
func (oc *Client) GetCurrentModel() string {
	return oc.model
}

// GetAvailableModels gets all available models from Ollama
func (oc *Client) GetAvailableModels() ([]string, error) {
	resp, err := oc.client.Get(oc.baseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(response.Models))
	for i, model := range response.Models {
		models[i] = model.Name
	}

	return models, nil
}

// testModel tests if a specific model is available
func (oc *Client) testModel(model string) bool {
	// Create a quick test client with shorter timeout
	testClient := &http.Client{Timeout: 5 * time.Second}

	reqBody := models.OllamaRequest{
		Model:  model,
		Prompt: "Hello",
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": 10, // Very short response for testing
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal test request for model %s: %v", model, err)
		return false
	}

	resp, err := testClient.Post(oc.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Test request failed for model %s: %v", model, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GenerateResponse generates a response using Ollama or fallback
func (oc *Client) GenerateResponse(prompt string) (string, error) {
	log.Printf("[DEBUG] GenerateResponse called with model: %s", oc.model)

	// First check if Ollama is available
	if !oc.TestConnection() {
		log.Printf("[DEBUG] Ollama not available, using fallback")
		return oc.generateFallbackResponse(prompt), nil
	}

	log.Printf("[DEBUG] Testing model availability: %s", oc.model)
	// Verify the current model is working, find alternative if not
	if !oc.testModel(oc.model) {
		log.Printf("[DEBUG] Model %s not working, searching for alternatives", oc.model)
		available, newModel := oc.FindAvailableModel()
		if !available {
			log.Printf("[DEBUG] No models available, using fallback")
			return oc.generateFallbackResponse(prompt), nil
		}
		log.Printf("[DEBUG] Switching to model: %s", newModel)
		oc.model = newModel
	}

	log.Printf("[DEBUG] Using model: %s for generation", oc.model)

	reqBody := models.OllamaRequest{
		Model:  oc.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": 1000, // Increased limit for more complete responses
			"temperature": 0.7,  // Reduce randomness for more focused responses
			"top_p":       0.9,  // Use nucleus sampling for better quality
		},
	}

	log.Printf("[DEBUG] Request body created for model: %s", reqBody.Model)
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[DEBUG] Failed to marshal request: %v", err)
		return oc.generateFallbackResponse(prompt), nil
	}

	log.Printf("[DEBUG] Sending POST request to: %s", oc.baseURL+"/api/generate")
	resp, err := oc.client.Post(oc.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[DEBUG] Ollama request failed: %v, using fallback", err)
		return oc.generateFallbackResponse(prompt), nil
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] Received response with status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[DEBUG] Ollama returned status %d, using fallback", resp.StatusCode)
		return oc.generateFallbackResponse(prompt), nil
	}

	var ollamaResp models.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Printf("[DEBUG] Failed to decode Ollama response: %v, using fallback", err)
		return oc.generateFallbackResponse(prompt), nil
	}

	log.Printf("[DEBUG] Response received, length: %d characters", len(ollamaResp.Response))
	// Check if response is empty
	if strings.TrimSpace(ollamaResp.Response) == "" {
		log.Printf("[DEBUG] Empty response received, using fallback")
		return oc.generateFallbackResponse(prompt), nil
	}

	// Add model signature to response
	response := strings.TrimSpace(ollamaResp.Response)
	response += fmt.Sprintf("\n\n---\n*Response generated by %s*", oc.model)
	log.Printf("[DEBUG] Successfully generated response using model: %s", oc.model)

	return response, nil
}

// generateFallbackResponse generates a fallback response when Ollama is unavailable
func (oc *Client) generateFallbackResponse(prompt string) string {
	// Extract user input from prompt
	lines := strings.Split(prompt, "\n")
	userIssue := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "User Issue:") {
			userIssue = strings.TrimSpace(strings.TrimPrefix(line, "User Issue:"))
			break
		}
	}

	lowerIssue := strings.ToLower(userIssue)

	// Generate basic troubleshooting based on keywords
	if strings.Contains(lowerIssue, "e1001") || strings.Contains(lowerIssue, "communication") || strings.Contains(lowerIssue, "timeout") {
		return `## ISSUE ANALYSIS:
- **Problem**: Communication timeout with device (Error E1001)
- **Likely causes**: Cable connection issues, power supply problems, or communication settings

## STEP-BY-STEP TROUBLESHOOTING:

1. **Check Physical Connections**
   - Verify all cable connections are secure
   - Inspect cables for damage or wear
   - Ensure proper cable routing away from interference sources

2. **Verify Power Supply**
   - Check that all devices have proper power
   - Verify voltage levels are within specification
   - Test power supply connections

3. **Test Communication Settings**
   - Verify baud rate and protocol settings
   - Check communication interface configuration
   - Test with known good settings

4. **Restart Communication Interface**
   - Power cycle the communication interface
   - Restart the LSIE application
   - Re-establish communication link

## ADDITIONAL RECOMMENDATIONS:
- **Preventive measures**: Regular cable inspection and connection maintenance
- **Escalate to support if**: Issue persists after all steps or multiple devices affected

*Note: AI assistant offline - using built-in troubleshooting knowledge*

---
*Response generated by Built-in Knowledge Base (Ollama offline)*`
	}

	if strings.Contains(lowerIssue, "e2005") || strings.Contains(lowerIssue, "temperature") || strings.Contains(lowerIssue, "sensor") {
		return `## ISSUE ANALYSIS:
- **Problem**: Temperature sensor fault (Error E2005)
- **Likely causes**: Sensor wiring issues, calibration problems, or faulty sensor

## STEP-BY-STEP TROUBLESHOOTING:

1. **Check Sensor Wiring**
   - Inspect all sensor connections
   - Look for loose or corroded connections
   - Verify proper wire routing and shielding

2. **Verify Sensor Calibration**
   - Check calibration settings in LSIE
   - Compare readings with known reference
   - Recalibrate if necessary

3. **Test Sensor Functionality**
   - Use multimeter to check sensor resistance
   - Verify sensor specifications
   - Test with known good sensor if available

4. **Update Configuration**
   - Check sensor configuration settings
   - Verify temperature ranges and limits
   - Update sensor parameters if needed

## ADDITIONAL RECOMMENDATIONS:
- **Preventive measures**: Regular sensor calibration and maintenance schedule
- **Escalate to support if**: Sensor replacement needed or configuration issues persist

*Note: AI assistant offline - using built-in troubleshooting knowledge*

---
*Response generated by Built-in Knowledge Base (Ollama offline)*`
	}

	if strings.Contains(lowerIssue, "e3010") || strings.Contains(lowerIssue, "power") || strings.Contains(lowerIssue, "voltage") {
		return `## ISSUE ANALYSIS:
- **Problem**: Power supply voltage out of range (Error E3010)
- **Likely causes**: Power supply failure, load issues, or configuration problems

## STEP-BY-STEP TROUBLESHOOTING:

1. **Check Input Voltage Levels**
   - Measure input voltage with multimeter
   - Verify voltage is within specification
   - Check for voltage fluctuations

2. **Inspect Power Supply Connections**
   - Verify all power connections are secure
   - Look for signs of overheating or damage
   - Check fuse and circuit breaker status

3. **Verify Load Requirements**
   - Calculate total power consumption
   - Ensure power supply capacity is adequate
   - Check for short circuits or overloads

4. **Replace Power Supply if Necessary**
   - Test with known good power supply
   - Check power supply specifications
   - Replace if output is out of tolerance

## ADDITIONAL RECOMMENDATIONS:
- **Preventive measures**: Regular power supply monitoring and maintenance
- **Escalate to support if**: Power supply replacement needed or electrical issues detected

*Note: AI assistant offline - using built-in troubleshooting knowledge*

---
*Response generated by Built-in Knowledge Base (Ollama offline)*`
	}

	if strings.Contains(lowerIssue, "cycler") || strings.Contains(lowerIssue, "limit") || strings.Contains(lowerIssue, "range") {
		return `## ISSUE ANALYSIS:
- **Problem**: Cycler limit out of range
- **Likely causes**: Configuration limits exceeded, calibration drift, or equipment malfunction

## STEP-BY-STEP TROUBLESHOOTING:

1. **Check Current Settings**
   - Review cycler voltage/current limits in software
   - Verify configured ranges match equipment specifications
   - Check for recent configuration changes

2. **Verify Equipment Status**
   - Inspect cycler for error indicators
   - Check all connections to the cycler
   - Ensure proper grounding and shielding

3. **Review Test Parameters**
   - Verify test voltage/current requirements
   - Check if limits are appropriate for the test
   - Confirm cell/battery specifications

4. **Calibration Check**
   - Verify last calibration date
   - Check calibration certificates
   - Perform basic calibration verification if needed

5. **Adjust Limits if Necessary**
   - Increase limits if within safe operating range
   - Ensure limits don't exceed equipment capabilities
   - Document any changes made

## ADDITIONAL RECOMMENDATIONS:
- **Preventive measures**: Regular calibration and limit verification
- **Escalate to support if**: Equipment shows signs of malfunction or calibration issues persist

*Note: AI assistant offline - using built-in troubleshooting knowledge*

---
*Response generated by Built-in Knowledge Base (Ollama offline)*`
	}

	// Generic response for other issues
	return fmt.Sprintf(`## ISSUE ANALYSIS:
- **Problem**: %s
- **Status**: AI assistant is currently offline, providing basic troubleshooting guidance

## GENERAL TROUBLESHOOTING STEPS:

1. **Document the Issue**
   - Note exact error messages
   - Record when the issue occurs
   - Document any recent changes

2. **Check Basic Connections**
   - Verify all cables and connections
   - Ensure proper power to all devices
   - Check for loose connections

3. **Review System Status**
   - Check system logs for errors
   - Verify all components are operational
   - Look for warning indicators

4. **Restart Components**
   - Power cycle affected devices
   - Restart software applications
   - Re-establish connections

## NEXT STEPS:
- Install and configure Ollama with gemma3:1b model for full AI assistance
- Consult technical documentation for specific error codes
- Contact support if issue persists

*Note: For full AI-powered troubleshooting, please install Ollama and run 'ollama pull llama3.2:1b' (recommended)*

---
*Response generated by Built-in Knowledge Base (Ollama offline)*`, userIssue)
}
