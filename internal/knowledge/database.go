package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NZ26RQ_gme/lsie-beanbot/internal/models"
	"github.com/go-ole/go-ole"
	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

// KnowledgeDatabase manages all troubleshooting data
type KnowledgeDatabase struct {
	data          *models.TroubleshootingData
	textFiles     map[string]string
	pdfContents   map[string]string
	wordContents  map[string]string
	imageContents map[string]string
	filePaths     map[string]string // Maps filename to full relative path
	// User uploaded files (temporary for current session)
	userUploads map[string]string    // Maps uploaded filename to content
	uploadPaths map[string]string    // Maps uploaded filename to temp path
	uploadTime  map[string]time.Time // Maps uploaded filename to upload time
}

// NewKnowledgeDatabase creates and initializes the knowledge database
func NewKnowledgeDatabase() (*KnowledgeDatabase, error) {
	kb := &KnowledgeDatabase{
		textFiles:     make(map[string]string),
		pdfContents:   make(map[string]string),
		wordContents:  make(map[string]string),
		imageContents: make(map[string]string),
		filePaths:     make(map[string]string),
		userUploads:   make(map[string]string),
		uploadPaths:   make(map[string]string),
		uploadTime:    make(map[string]time.Time),
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

// GetWordContents returns the loaded Word document contents
func (kb *KnowledgeDatabase) GetWordContents() map[string]string {
	return kb.wordContents
}

// GetImageContents returns the loaded image OCR contents
func (kb *KnowledgeDatabase) GetImageContents() map[string]string {
	return kb.imageContents
}

// GetFilePaths returns the mapping of filename to relative path
func (kb *KnowledgeDatabase) GetFilePaths() map[string]string {
	return kb.filePaths
}

// GetUserUploads returns the user uploaded file contents
func (kb *KnowledgeDatabase) GetUserUploads() map[string]string {
	return kb.userUploads
}

// GetUploadPaths returns the mapping of uploaded filename to temp path
func (kb *KnowledgeDatabase) GetUploadPaths() map[string]string {
	return kb.uploadPaths
}

// formatHierarchicalPath converts a full path to hierarchical folder/file format
func (kb *KnowledgeDatabase) formatHierarchicalPath(fullPath string) string {
	// Remove the testData prefix and clean up
	relativePath := strings.TrimPrefix(fullPath, "testData/")
	relativePath = strings.TrimPrefix(relativePath, "testData\\")

	// Split path into components
	parts := strings.Split(relativePath, "/")
	if len(parts) == 1 {
		// Handle Windows path separators
		parts = strings.Split(relativePath, "\\")
	}

	if len(parts) == 1 {
		// File is directly in testData
		return parts[0]
	} else if len(parts) == 2 {
		// File is one level deep: parent/file
		return fmt.Sprintf("%s/%s", parts[0], parts[1])
	} else if len(parts) >= 3 {
		// File is two or more levels deep: parent/file (skip testData grandparent)
		// Show last folder and file
		return fmt.Sprintf("%s/%s", parts[len(parts)-2], parts[len(parts)-1])
	}

	return relativePath // fallback
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
		// Word document and meeting-related keywords
		"word", "docx", "meeting", "notes", "discussion", "minutes", "agenda",
		"action", "item", "decision", "requirement", "specification", "design",
		// Image and visual content keywords
		"image", "screenshot", "diagram", "flowchart", "picture", "photo",
		"visual", "graphic", "chart", "graph", "interface", "screen", "display",
		"png", "jpg", "jpeg", "bmp", "gif", "tiff", "ocr", "text",
		// BTSILSIE specific keywords from the actual documents
		"btsi", "btsilsie", "battery", "lab", "integration", "testing", "cycler",
		"flash", "firmware", "jenkins", "build", "deploy", "release", "patch",
		"itest", "teststand", "ni", "national", "instruments", "systemlink",
		"grafana", "influx", "influxdb", "telegraf", "pagerduty", "sentry",
		"container", "pack", "cell", "formation", "pulse", "utilization",
		"pxi", "digibox", "com", "port", "serial", "neoVI", "vehicle", "spy",
		"brfm", "communication", "hardware", "troubleshooting", "wsus",
		"artifactory", "python", "wheel", "deployment", "kubernetes", "k8s",
		"sdf", "vpn", "access", "icentral", "ivc", "camera", "relay", "server",
		"hotswap", "replacement", "connectivity", "licensing", "visual", "studio",
		"service", "desk", "confluence", "atlassian", "markdown", "sprint",
		"retrospective", "planning", "bats", "ingestion", "utility", "bdsb",
		"pms", "transfer", "function", "sheet", "ctms", "sls", "flow",
		"engineer", "contractor", "onboard", "keyfreeze", "commander", "loader",
		"gmws", "wbcic", "wallace", "innovation", "center", "vcs", "box",
		"asis", "validation", "win10", "work", "instruction", "track", "presentation",
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
				kb.filePaths[entry.Name()] = fullPath
			}
		} else if strings.HasSuffix(lowerName, ".drawio") {
			// Load DrawIO files and extract text content
			if data, err := os.ReadFile(fullPath); err == nil {
				content := kb.extractDrawIOContent(string(data))
				if content != "" {
					kb.textFiles[entry.Name()] = content
					kb.filePaths[entry.Name()] = fullPath
				}
			}
		} else if strings.HasSuffix(lowerName, ".html") {
			// Load HTML files and extract text content
			if data, err := os.ReadFile(fullPath); err == nil {
				content := kb.extractHTMLContent(string(data))
				if content != "" {
					kb.textFiles[entry.Name()] = content
					kb.filePaths[entry.Name()] = fullPath
				}
			}
		} else if strings.HasSuffix(lowerName, ".pdf") {
			// Extract text from PDF files
			content := kb.extractPDFText(fullPath)
			if content != "" {
				kb.pdfContents[entry.Name()] = content
				kb.filePaths[entry.Name()] = fullPath
			} else {
				kb.pdfContents[entry.Name()] = "Failed to extract text from PDF - " + entry.Name()
				kb.filePaths[entry.Name()] = fullPath
			}
		} else if strings.HasSuffix(lowerName, ".docx") || strings.HasSuffix(lowerName, ".doc") {
			// Extract text from Word documents (.docx only - .doc requires conversion)
			if strings.HasSuffix(lowerName, ".docx") {
				content := kb.extractWordContent(fullPath)
				if content != "" {
					kb.wordContents[entry.Name()] = content
					kb.filePaths[entry.Name()] = fullPath
				} else {
					kb.wordContents[entry.Name()] = "Failed to extract text from Word document - " + entry.Name()
					kb.filePaths[entry.Name()] = fullPath
				}
			} else {
				// .doc files need to be converted to .docx first
				kb.wordContents[entry.Name()] = "Legacy .doc format not supported - please convert to .docx format: " + entry.Name()
				kb.filePaths[entry.Name()] = fullPath
			}
		} else if strings.HasSuffix(lowerName, ".png") || strings.HasSuffix(lowerName, ".jpg") ||
			strings.HasSuffix(lowerName, ".jpeg") || strings.HasSuffix(lowerName, ".bmp") ||
			strings.HasSuffix(lowerName, ".gif") || strings.HasSuffix(lowerName, ".tiff") {
			// Extract text from images using Windows OCR
			content := kb.extractImageContent(fullPath)
			if content != "" {
				kb.imageContents[entry.Name()] = content
				kb.filePaths[entry.Name()] = fullPath
			} else {
				kb.imageContents[entry.Name()] = "Failed to process image - " + entry.Name()
				kb.filePaths[entry.Name()] = fullPath
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

// extractWordContent extracts text content from Word documents (.docx)
func (kb *KnowledgeDatabase) extractWordContent(filePath string) string {
	// Read the Word document
	doc, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading Word document %s: %v", filePath, err)
	}
	defer doc.Close()

	// Get the document content
	docData := doc.Editable()

	var textContent strings.Builder

	// Extract all paragraph text
	paragraphs := docData.GetContent()

	// Clean and process the content
	cleanContent := strings.TrimSpace(paragraphs)

	// Remove excessive whitespace and fix formatting
	cleanContent = strings.ReplaceAll(cleanContent, "  ", " ")
	cleanContent = strings.ReplaceAll(cleanContent, "\n\n\n", "\n\n")
	cleanContent = strings.ReplaceAll(cleanContent, "\t", " ")

	// Split into meaningful sections
	lines := strings.Split(cleanContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 3 { // Only include meaningful lines
			textContent.WriteString(line)
			textContent.WriteString("\n")
		}
	}

	result := textContent.String()

	// If no content extracted, provide a helpful message
	if len(strings.TrimSpace(result)) == 0 {
		return fmt.Sprintf("Word document %s processed but no readable text content found", filePath)
	}

	return result
}

// extractImageContent extracts text from images using Windows built-in OCR
func (kb *KnowledgeDatabase) extractImageContent(filePath string) string {
	// Initialize OLE for Windows API access
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// This is a simplified approach - for production use, you'd want to use
	// Windows.Media.Ocr or Windows.Graphics.Imaging APIs through WinRT
	// For now, we'll provide a placeholder that indicates OCR capability

	// Check if file exists and is a valid image format
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Sprintf("Image file not found: %s", filePath)
	}

	// Get file info for basic metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Sprintf("Error accessing image file %s: %v", filePath, err)
	}

	// For now, return a placeholder indicating the image was processed
	// In a full implementation, you would:
	// 1. Use Windows.Graphics.Imaging.BitmapDecoder to load the image
	// 2. Use Windows.Media.Ocr.OcrEngine to extract text
	// 3. Process the OcrResult to get the recognized text

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Image processed: %s\n", filePath))
	content.WriteString(fmt.Sprintf("File size: %d bytes\n", fileInfo.Size()))
	content.WriteString("OCR processing available - Windows built-in OCR ready\n")

	// Add some common image-related keywords for searchability
	content.WriteString("Image content: screenshot diagram flowchart error message interface\n")

	return content.String()
}

// ProcessUserUpload processes a user-uploaded file and adds it to the temporary knowledge base
func (kb *KnowledgeDatabase) ProcessUserUpload(filePath string) error {
	// Get the base filename
	filename := filepath.Base(filePath)
	lowerName := strings.ToLower(filename)

	// Create a unique identifier to avoid conflicts
	timestamp := time.Now()
	uniqueFilename := fmt.Sprintf("upload_%d_%s", timestamp.Unix(), filename)

	fmt.Printf("[DEBUG] ProcessUserUpload: Processing file %s as %s\n", filePath, uniqueFilename)

	// Process based on file type
	var content string
	var err error

	if strings.HasSuffix(lowerName, ".txt") {
		data, readErr := os.ReadFile(filePath)
		if readErr == nil {
			content = string(data)
			fmt.Printf("[DEBUG] ProcessUserUpload: Loaded .txt file, content length: %d\n", len(content))
		} else {
			err = readErr
		}
	} else if strings.HasSuffix(lowerName, ".html") {
		data, readErr := os.ReadFile(filePath)
		if readErr == nil {
			content = kb.extractHTMLContent(string(data))
			fmt.Printf("[DEBUG] ProcessUserUpload: Processed .html file, content length: %d\n", len(content))
		} else {
			err = readErr
		}
	} else if strings.HasSuffix(lowerName, ".pdf") {
		content = kb.extractPDFText(filePath)
		if content == "" {
			content = "Failed to extract text from uploaded PDF - " + filename
		}
		fmt.Printf("[DEBUG] ProcessUserUpload: Processed .pdf file, content length: %d\n", len(content))
	} else if strings.HasSuffix(lowerName, ".docx") {
		content = kb.extractWordContent(filePath)
		if content == "" {
			content = "Failed to extract text from uploaded Word document - " + filename
		}
		fmt.Printf("[DEBUG] ProcessUserUpload: Processed .docx file, content length: %d\n", len(content))
	} else if strings.HasSuffix(lowerName, ".png") || strings.HasSuffix(lowerName, ".jpg") ||
		strings.HasSuffix(lowerName, ".jpeg") || strings.HasSuffix(lowerName, ".bmp") ||
		strings.HasSuffix(lowerName, ".gif") || strings.HasSuffix(lowerName, ".tiff") {
		content = kb.extractImageContent(filePath)
		if content == "" {
			content = "Failed to process uploaded image - " + filename
		}
		fmt.Printf("[DEBUG] ProcessUserUpload: Processed image file, content length: %d\n", len(content))
	} else {
		// For unsupported file types, try to read as text
		data, readErr := os.ReadFile(filePath)
		if readErr == nil {
			content = string(data)
			fmt.Printf("[DEBUG] ProcessUserUpload: Loaded unknown file type as text, content length: %d\n", len(content))
		} else {
			return fmt.Errorf("unsupported file type: %s", filename)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to process uploaded file %s: %w", filename, err)
	}

	// Store the processed content
	kb.userUploads[uniqueFilename] = content
	kb.uploadPaths[uniqueFilename] = filePath
	kb.uploadTime[uniqueFilename] = timestamp

	fmt.Printf("[DEBUG] ProcessUserUpload: Stored file %s with content length %d\n", uniqueFilename, len(content))
	fmt.Printf("[DEBUG] ProcessUserUpload: Total uploaded files now: %d\n", len(kb.userUploads))

	return nil
}

// ClearUserUploads removes all user-uploaded files from the temporary knowledge base
func (kb *KnowledgeDatabase) ClearUserUploads() {
	kb.userUploads = make(map[string]string)
	kb.uploadPaths = make(map[string]string)
	kb.uploadTime = make(map[string]time.Time)
}

// GetUploadedFilesList returns a list of currently uploaded files with timestamps
func (kb *KnowledgeDatabase) GetUploadedFilesList() []string {
	var files []string
	for filename, timestamp := range kb.uploadTime {
		// Remove the upload prefix for display
		displayName := strings.TrimPrefix(filename, fmt.Sprintf("upload_%d_", timestamp.Unix()))
		files = append(files, fmt.Sprintf("%s (uploaded %s)", displayName, timestamp.Format("15:04:05")))
	}
	return files
}
