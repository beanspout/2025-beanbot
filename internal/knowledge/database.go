package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NZ26RQ_gme/lsie-beanbot/internal/models"
	"github.com/ledongthuc/pdf"
)

// KnowledgeDatabase manages all troubleshooting data
type KnowledgeDatabase struct {
	data        *models.TroubleshootingData
	textFiles   map[string]string
	pdfContents map[string]string
}

// NewKnowledgeDatabase creates and initializes the knowledge database
func NewKnowledgeDatabase() (*KnowledgeDatabase, error) {
	kb := &KnowledgeDatabase{
		textFiles:   make(map[string]string),
		pdfContents: make(map[string]string),
	}

	// Load JSON data
	jsonData, err := os.ReadFile("testData/lsie_errors.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read lsie_errors.json: %w", err)
	}

	if err := json.Unmarshal(jsonData, &kb.data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	// Load all text files from testData directory
	kb.loadTextFiles("testData")

	return kb, nil
}

// GetData returns the troubleshooting data
func (kb *KnowledgeDatabase) GetData() *models.TroubleshootingData {
	return kb.data
}

// GetTextFiles returns the loaded text files
func (kb *KnowledgeDatabase) GetTextFiles() map[string]string {
	return kb.textFiles
}

// GetPDFContents returns the loaded PDF contents
func (kb *KnowledgeDatabase) GetPDFContents() map[string]string {
	return kb.pdfContents
}

// ContainsAnyKeyword checks if input contains any of the keywords
func (kb *KnowledgeDatabase) ContainsAnyKeyword(input string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(input, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// IsRelevantContent determines if text content is relevant to the user input
func (kb *KnowledgeDatabase) IsRelevantContent(userInput, content string) bool {
	lowerContent := strings.ToLower(content)
	lowerInput := strings.ToLower(userInput)

	// Check for direct keyword matches
	keywords := strings.Fields(lowerInput)
	relevantKeywords := 0

	for _, keyword := range keywords {
		if len(keyword) > 2 && strings.Contains(lowerContent, keyword) { // Lowered threshold from 3 to 2
			relevantKeywords++
		}
	}

	// Enhanced keyword matching for troubleshooting content and comprehensive LSIE documentation
	troubleshootingKeywords := []string{
		"error", "troubleshoot", "communication", "sensor", "power", "temperature",
		"timeout", "connection", "voltage", "calibration", "cycler", "device",
		"interface", "problem", "issue", "solution", "step", "procedure",
		"check", "verify", "test", "replace", "restart", "configure",
		"lsie", "support", "jira", "ticket", "contact", "help", "official",
		"execution", "standard", "process", "team", "unofficial", "pdf", "file",
		"open", "document", "manual", "guide", "instruction", "setup", "install",
		"software", "hardware", "system", "application", "program", "tool",
		// LSIE specific keywords
		"lsie", "solutionbuilder", "testmanager", "automation", "python", "vcl",
		"channel", "module", "schedule", "display", "data", "logging", "report",
		"security", "configuration", "developer", "api", "scripting", "control",
		"panel", "limit", "alarm", "calculation", "variable", "function",
		"installation", "getting", "started", "how", "use", "managing", "creating",
	}

	keywordMatches := 0
	for _, keyword := range troubleshootingKeywords {
		if strings.Contains(lowerInput, keyword) && strings.Contains(lowerContent, keyword) {
			keywordMatches++
		}
	}

	// More inclusive matching - return true if ANY of these conditions are met:
	// 1. At least 1 relevant keyword match (instead of 2)
	// 2. Any troubleshooting keyword matches
	// 3. If user input is short (< 10 chars), include content more liberally
	// 4. Contains general troubleshooting terms
	return relevantKeywords >= 1 ||
		keywordMatches > 0 ||
		len(lowerInput) < 10 ||
		strings.Contains(lowerContent, "troubleshoot") ||
		strings.Contains(lowerContent, "solution") ||
		strings.Contains(lowerContent, "procedure") ||
		(strings.Contains(lowerContent, "error") && len(lowerContent) > 50)
}

// extractDrawIOContent extracts text content from DrawIO XML
func (kb *KnowledgeDatabase) extractDrawIOContent(xmlContent string) string {
	var content strings.Builder

	// Look for value attributes which contain the text content
	// Simple extraction - look for value="..." patterns
	lines := strings.Split(xmlContent, "\n")
	for _, line := range lines {
		// Look for value attributes in XML
		if strings.Contains(line, "value=") {
			// Extract text between value="..."
			start := strings.Index(line, `value="`)
			if start != -1 {
				start += 7 // Skip 'value="'
				end := strings.Index(line[start:], `"`)
				if end != -1 {
					text := line[start : start+end]
					// Decode HTML entities and clean up
					text = strings.ReplaceAll(text, "&quot;", "\"")
					text = strings.ReplaceAll(text, "&amp;", "&")
					text = strings.ReplaceAll(text, "&lt;", "<")
					text = strings.ReplaceAll(text, "&gt;", ">")
					text = strings.ReplaceAll(text, "&#xa;", "\n")

					// Only include meaningful text (not single chars or very short)
					if len(strings.TrimSpace(text)) > 5 {
						content.WriteString(text + "\n")
					}
				}
			}
		}
	}

	return content.String()
}

// extractHTMLContent extracts text content from HTML (basic implementation)
func (kb *KnowledgeDatabase) extractHTMLContent(htmlContent string) string {
	var content strings.Builder

	// Very basic HTML text extraction
	// Look for content between tags that might contain useful text
	lines := strings.Split(htmlContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and common HTML tags
		if line == "" || strings.HasPrefix(line, "<!") ||
			strings.HasPrefix(line, "<html") || strings.HasPrefix(line, "<head") ||
			strings.HasPrefix(line, "<meta") || strings.HasPrefix(line, "<link") ||
			strings.HasPrefix(line, "<script") || strings.HasPrefix(line, "<style") {
			continue
		}

		// Extract title content
		if strings.Contains(line, "<title>") && strings.Contains(line, "</title>") {
			start := strings.Index(line, "<title>") + 7
			end := strings.Index(line, "</title>")
			if start < end {
				title := line[start:end]
				if len(strings.TrimSpace(title)) > 0 {
					content.WriteString("Title: " + title + "\n")
				}
			}
		}

		// Look for any text content that might be embedded
		// This is a simple approach - in reality, you'd want proper HTML parsing
		if strings.Contains(line, "troubleshoot") || strings.Contains(line, "error") ||
			strings.Contains(line, "problem") || strings.Contains(line, "solution") ||
			strings.Contains(line, "step") || strings.Contains(line, "issue") {
			// Try to extract meaningful text
			cleaned := strings.ReplaceAll(line, "&quot;", "\"")
			cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
			cleaned = strings.ReplaceAll(cleaned, "\\n", "\n")
			if len(cleaned) > 20 && len(cleaned) < 500 {
				content.WriteString(cleaned + "\n")
			}
		}
	}

	return content.String()
}

// loadTextFiles recursively loads all text files from a directory
func (kb *KnowledgeDatabase) loadTextFiles(dirPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := dirPath + "/" + entry.Name()
		lowerName := strings.ToLower(entry.Name())

		if entry.IsDir() {
			// Recursively load from subdirectories
			kb.loadTextFiles(fullPath)
		} else if strings.HasSuffix(lowerName, ".txt") {
			// Load text files
			if data, err := os.ReadFile(fullPath); err == nil {
				kb.textFiles[entry.Name()] = string(data)
			}
		} else if strings.HasSuffix(lowerName, ".drawio") {
			// Load DrawIO files and extract text content
			if data, err := os.ReadFile(fullPath); err == nil {
				content := kb.extractDrawIOContent(string(data))
				if content != "" {
					kb.textFiles[entry.Name()] = content
				}
			}
		} else if strings.HasSuffix(lowerName, ".html") {
			// Load HTML files and extract text content
			if data, err := os.ReadFile(fullPath); err == nil {
				content := kb.extractHTMLContent(string(data))
				if content != "" {
					kb.textFiles[entry.Name()] = content
				}
			}
		} else if strings.HasSuffix(lowerName, ".pdf") {
			// Extract text from PDF files
			content := kb.extractPDFText(fullPath)
			if content != "" {
				kb.pdfContents[entry.Name()] = content
			} else {
				kb.pdfContents[entry.Name()] = "Failed to extract text from PDF - " + entry.Name()
			}
		}
	}
}

// extractPDFText extracts text content from a PDF file
func (kb *KnowledgeDatabase) extractPDFText(filePath string) string {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var textContent strings.Builder
	numPages := reader.NumPage()

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Extract text from the page using correct API
		fonts := make(map[string]*pdf.Font)
		text, err := page.GetPlainText(fonts)
		if err != nil {
			continue
		}

		// Clean up the text and fix encoding issues
		cleanText := strings.TrimSpace(text)

		// Fix common PDF encoding issues
		cleanText = strings.ReplaceAll(cleanText, "♥", " ")
		cleanText = strings.ReplaceAll(cleanText, "◄", " ")
		cleanText = strings.ReplaceAll(cleanText, "↔", " ")
		cleanText = strings.ReplaceAll(cleanText, "�", " ")

		// Remove excessive whitespace
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
		cleanText = strings.ReplaceAll(cleanText, "\n\n\n", "\n\n")

		if cleanText != "" && len(cleanText) > 10 {
			textContent.WriteString(cleanText)
			textContent.WriteString("\n")
		}
	}

	result := textContent.String()
	return result
}
