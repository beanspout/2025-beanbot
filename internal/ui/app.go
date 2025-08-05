package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/NZ26RQ_gme/lsie-beanbot/internal/knowledge"
	"github.com/NZ26RQ_gme/lsie-beanbot/internal/ollama"
)

// BeanBot represents the main application UI structure
type BeanBot struct {
	app             fyne.App
	window          fyne.Window
	knowledgeDB     *knowledge.KnowledgeDatabase
	ollamaClient    *ollama.Client
	submitBtn       *widget.Button
	statusLabel     *widget.Label     // Add reference to status label for updates
	debugMode       bool              // Debug mode flag
	scrollContainer *container.Scroll // Add reference to scroll container
}

// NewBeanBot creates a new BeanBot UI instance with all required dependencies
func NewBeanBot(app fyne.App, window fyne.Window, kb *knowledge.KnowledgeDatabase, client *ollama.Client) *BeanBot {
	return &BeanBot{
		app:          app,
		window:       window,
		knowledgeDB:  kb,
		ollamaClient: client,
	}
}

// SetupUI sets up the main UI
func (b *BeanBot) SetupUI() {
	// Create main layout without header since it's redundant
	content := container.NewBorder(
		nil,                   // No header - window title is sufficient
		b.createFooter(),      // Footer with cute status
		nil,                   // No left sidebar
		nil,                   // No right sidebar
		b.createMainContent(), // Main content area
	)

	b.window.SetContent(content)
}

// createHeader creates the header section
func (b *BeanBot) createHeader() *fyne.Container {
	// No header needed - window title and footer status are sufficient
	return nil
}

// createFooter creates the footer section
func (b *BeanBot) createFooter() *fyne.Container {
	status := widget.NewLabelWithStyle("🤖 BeanBot AI ⏳ warming up...",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Store reference to status label
	b.statusLabel = status

	// Make status clickable for model selection
	statusButton := widget.NewButton("", func() {
		b.showModelSelectionDialog()
	})
	statusButton.Importance = widget.LowImportance

	// Create a container that overlays the button on the label
	statusContainer := container.NewStack(status, statusButton)

	// Test Ollama connection
	go func() {
		b.debugLog("Testing Ollama connection...")
		if b.ollamaClient.TestConnection() {
			b.debugLog("Ollama connection successful, searching for available models")
			// Try to get available models (preferring llama3.2:1b)
			available, model := b.ollamaClient.FindAvailableModel()
			if available {
				b.debugLog("Found available model: %s", model)
				status.SetText(fmt.Sprintf("🤖 BeanBot AI - %s ✅ ready to help! (click to change)", model))
				// Update the client to use the found model
				b.ollamaClient.SetModel(model)
				b.debugLog("Set active model to: %s", model)
			} else {
				b.debugLog("No models found")
				status.SetText("🤖 BeanBot AI ❌ no models found - install with: ollama pull llama3.2:1b")
			}
		} else {
			b.debugLog("Ollama connection failed - server offline")
			status.SetText("🤖 BeanBot AI ❌ offline")
		}
	}()

	return container.NewVBox(
		widget.NewSeparator(),
		statusContainer,
	)
}

// createMainContent creates the main content area
func (b *BeanBot) createMainContent() fyne.CanvasObject {
	// Create input area components - clean chat interface without labels
	inputEntry := widget.NewMultiLineEntry()
	inputEntry.SetPlaceHolder("Describe your issue, ... (Example: I'm having trouble with system communication or getting an error message, hint: Upload screenshots, log files, or error reports so I can get more context on your problem)")
	inputEntry.Wrapping = fyne.TextWrapWord   // Enable word wrapping for input too
	inputEntry.Resize(fyne.NewSize(800, 100)) // Set a reasonable height for input

	// Create response area using RichText with Border Layout Pattern (no white space)
	responseText := widget.NewRichTextFromMarkdown("\n\n\n\n## 🤖 Hi there! \n\n### What engineering challenge can I help you with today? 💭")
	responseText.Wrapping = fyne.TextWrapWord

	// Create submit button with handler for RichText
	submitBtn := widget.NewButton("Ask", func() {
		b.handleEngineeringRequest(inputEntry.Text, responseText)
	})
	submitBtn.Importance = widget.HighImportance

	// Create clear button to reset everything
	clearBtn := widget.NewButton("Clear", func() {
		// Clear the input field
		inputEntry.SetText("")
		// Clear user uploads
		b.knowledgeDB.ClearUserUploads()
		// Reset response area to welcome message
		responseText.ParseMarkdown("\n\n\n\n## 🤖 Hi there! \n\n### What engineering challenge can I help you with today? 💭")
		// Scroll to top when clearing
		if b.scrollContainer != nil {
			b.scrollContainer.ScrollToTop()
		}
	})

	// Create upload button to add user files
	uploadBtn := widget.NewButton("Upload Files", func() {
		b.handleFileUpload(responseText)
	})
	uploadBtn.Importance = widget.MediumImportance

	// Store reference to button for progress handling
	b.submitBtn = submitBtn

	// Fixed content for bottom section (input area) - clean chat-style layout with three buttons
	buttonContainer := container.NewGridWithColumns(3, submitBtn, uploadBtn, clearBtn)
	bottomSection := container.NewVBox(
		inputEntry,
		buttonContainer,
	)

	// Apply Border Layout Pattern to eliminate "big box" scroll container issue
	// Following the proven pattern: fixed content in bottom, scrollable content in center
	// Use border layout with spacers to center content horizontally
	leftSpacer := container.NewWithoutLayout()
	rightSpacer := container.NewWithoutLayout()
	centeredContent := container.NewBorder(nil, nil, leftSpacer, rightSpacer, responseText)

	// Create scroll container and store reference for programmatic scrolling
	scrollContainer := container.NewScroll(centeredContent)
	b.scrollContainer = scrollContainer

	mainContainer := container.NewBorder(
		nil,             // Top - not needed (no header)
		bottomSection,   // Bottom - fixed size input area (chat-style)
		nil,             // Left - not needed
		nil,             // Right - not needed
		scrollContainer, // Center - scrollable centered RichText (takes remaining space)
	)

	return mainContainer
}

// handleEngineeringRequest handles the engineering support request
func (b *BeanBot) handleEngineeringRequest(userInput string, responseEntry *widget.RichText) {
	if strings.TrimSpace(userInput) == "" {
		emptyResponse := "Please describe your engineering issue to get started."
		emptyResponse += "\n\n---\n\n**📚 Sources Referenced:**\n\n"
		emptyResponse += "*No documents from testData were referenced because no query was provided.*\n"
		responseEntry.ParseMarkdown(emptyResponse)
		return
	}

	b.debugLog("Handling engineering request: %s", userInput)
	b.debugLog("Current model: %s", b.ollamaClient.GetCurrentModel())

	// Scroll to top when Ask is pressed
	if b.scrollContainer != nil {
		b.scrollContainer.ScrollToTop()
		b.debugLog("Scrolled to top of response area")
	}

	// Show progress by changing button text
	originalText := b.submitBtn.Text
	b.submitBtn.SetText("Processing...")
	b.submitBtn.Disable()
	responseEntry.ParseMarkdown("\n\n\n\n## 🔍 Looking into this for you... \n\n### ✨ Just a moment! ✨")

	go func() {
		defer func() {
			// Restore button
			b.submitBtn.SetText(originalText)
			b.submitBtn.Enable()
		}()

		b.debugLog("Building engineering context...")
		// Build context from knowledge database
		context, sources := b.buildEngineeringContext(userInput)
		b.debugLog("Context length: %d characters", len(context))
		b.debugLog("Referenced %d source documents", len(sources))

		// Create prompt for Ollama
		prompt := b.createEngineeringPrompt(userInput, context)
		b.debugLog("Prompt length: %d characters", len(prompt))

		// Check if this is a direct response (not a prompt for Ollama)
		if strings.Contains(context, "outside my technical troubleshooting expertise") {
			b.debugLog("Using direct response (outside expertise)")
			// Always add source information even for direct responses
			context += "\n\n---\n\n**📚 Sources Referenced:**\n\n"
			if len(sources) > 0 {
				for i, source := range sources {
					context += fmt.Sprintf("%d. %s\n", i+1, source)
				}
			} else {
				context += "*No relevant documents from testData were found for this query. This response indicates the question is outside the scope of available technical documentation.*\n"
			}
			responseEntry.ParseMarkdown(context)
			return
		}

		b.debugLog("Sending request to Ollama with model: %s", b.ollamaClient.GetCurrentModel())
		// Get response from Ollama
		response, err := b.ollamaClient.GenerateResponse(prompt)
		if err != nil {
			b.debugLog("Error getting AI response: %v", err)
			log.Printf("Error getting AI response: %v", err)
			errorResponse := fmt.Sprintf("Error getting AI response: %v", err)
			// Always add source information even for error responses
			errorResponse += "\n\n---\n\n**📚 Sources Referenced:**\n\n"
			if len(sources) > 0 {
				for i, source := range sources {
					errorResponse += fmt.Sprintf("%d. %s\n", i+1, source)
				}
			} else {
				errorResponse += "*No documents from testData were referenced due to the error. Please try rephrasing your question.*\n"
			}
			responseEntry.ParseMarkdown(errorResponse)
			return
		}

		b.debugLog("Received response from Ollama, length: %d characters", len(response))

		// Always add source references to the response - this is mandatory
		response += "\n\n---\n\n**📚 Sources Referenced:**\n\n"
		if len(sources) > 0 {
			for i, source := range sources {
				response += fmt.Sprintf("%d. %s\n", i+1, source)
			}
		} else {
			response += "*No documents from testData were referenced for this response. This answer is based on general AI knowledge and may not reflect your specific documentation or procedures.*\n"
		}

		// Display response in the same window
		responseEntry.ParseMarkdown(response)
	}()
}

// handleFileUpload handles user file uploads using Windows system dialog
func (b *BeanBot) handleFileUpload(responseEntry *widget.RichText) {
	b.debugLog("Opening file upload dialog")

	// Show the system file dialog
	files, err := ShowFileDialog()
	if err != nil {
		b.debugLog("Error opening file dialog: %v", err)
		dialog.ShowError(fmt.Errorf("failed to open file dialog: %w", err), b.window)
		return
	}

	if len(files) == 0 {
		b.debugLog("No files selected")
		return // User cancelled
	}

	b.debugLog("Processing %d uploaded files", len(files))

	// Show processing message
	responseEntry.ParseMarkdown("\n\n\n\n## 📁 Processing uploaded files... \n\n### ✨ Please wait while I analyze your files ✨")

	// Process files in background
	go func() {
		var processedFiles []string
		var errors []string

		for _, filePath := range files {
			err := b.knowledgeDB.ProcessUserUpload(filePath)
			if err != nil {
				b.debugLog("Error processing file %s: %v", filePath, err)
				errors = append(errors, fmt.Sprintf("• %s: %v", filePath, err))
			} else {
				b.debugLog("Successfully processed file: %s", filePath)
				processedFiles = append(processedFiles, filePath)
			}
		}

		// Build response message
		var message strings.Builder
		message.WriteString("\n\n\n\n## 📁 File Upload Complete! \n\n")

		if len(processedFiles) > 0 {
			message.WriteString("### ✅ Successfully uploaded and processed:\n\n")
			for i, file := range processedFiles {
				fileName := filepath.Base(file)
				message.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, fileName))
			}
			message.WriteString("\n*These files are now available for your questions and will be included in AI responses.*\n\n")
		}

		if len(errors) > 0 {
			message.WriteString("### ❌ Failed to process:\n\n")
			for _, errMsg := range errors {
				message.WriteString(errMsg + "\n")
			}
			message.WriteString("\n")
		}

		// Show currently uploaded files
		uploadedList := b.knowledgeDB.GetUploadedFilesList()
		if len(uploadedList) > 0 {
			message.WriteString("### 📋 All uploaded files in this session:\n\n")
			for i, file := range uploadedList {
				message.WriteString(fmt.Sprintf("%d. %s\n", i+1, file))
			}
			message.WriteString("\n*Use 'Clear' button to remove uploaded files and start fresh.*\n")
		}

		responseEntry.ParseMarkdown(message.String())
	}()
}

// buildEngineeringContext builds context from the knowledge database and returns sources
func (b *BeanBot) buildEngineeringContext(userInput string) (string, []string) {
	var context strings.Builder
	var sources []string
	lowerInput := strings.ToLower(userInput)

	// PRIORITY 0: Include user-uploaded files first (highest priority)
	// User uploads get preferential treatment - include them more liberally since user specifically uploaded them
	userUploads := b.knowledgeDB.GetUserUploads()
	b.debugLog("Processing user uploads: found %d uploaded files", len(userUploads))

	for filename, content := range userUploads {
		b.debugLog("Checking uploaded file: %s, content length: %d", filename, len(content))

		// For user uploads, use much more liberal inclusion criteria
		// Include if ANY of these conditions are met:
		// 1. Contains any word from user input (even short words)
		// 2. User input is very short (general query - include all uploads)
		// 3. Contains common troubleshooting keywords
		// 4. File has substantial content (user uploaded it for a reason)
		shouldInclude := false

		if len(strings.TrimSpace(lowerInput)) <= 10 {
			// Very short queries - include all uploaded files
			shouldInclude = true
			b.debugLog("Including %s: short user query", filename)
		} else if len(content) > 50 {
			// Check for any word matches (much more liberal than IsRelevantContent)
			inputWords := strings.Fields(lowerInput)
			lowerContent := strings.ToLower(content)

			for _, word := range inputWords {
				if len(word) > 2 && strings.Contains(lowerContent, word) {
					shouldInclude = true
					b.debugLog("Including %s: found word match '%s'", filename, word)
					break
				}
			}

			// Also include if content has technical keywords
			technicalKeywords := []string{"error", "problem", "issue", "step", "solution", "configure", "install", "troubleshoot"}
			for _, keyword := range technicalKeywords {
				if strings.Contains(lowerContent, keyword) {
					shouldInclude = true
					b.debugLog("Including %s: contains technical keyword '%s'", filename, keyword)
					break
				}
			}
		}

		if shouldInclude {
			b.debugLog("File %s is included for user input", filename)
			// Remove timestamp prefix for display
			displayName := filename
			if strings.Contains(filename, "_") {
				parts := strings.SplitN(filename, "_", 3)
				if len(parts) >= 3 {
					displayName = parts[2] // Get the original filename part
				}
			}

			context.WriteString(fmt.Sprintf("From User Upload (%s):\n", displayName))
			sources = append(sources, "User Upload: "+displayName)
			// Give more content space to user uploads since they're specifically relevant
			if len(content) > 800 {
				context.WriteString(content[:800] + "...\n\n")
			} else {
				context.WriteString(content + "\n\n")
			}
		} else {
			b.debugLog("File %s is NOT included for user input '%s'", filename, lowerInput)
		}
	}

	// Check if this is a technical engineering question vs general question
	isTechnicalQuestion := strings.Contains(lowerInput, "error") ||
		strings.Contains(lowerInput, "problem") ||
		strings.Contains(lowerInput, "troubleshoot") ||
		strings.Contains(lowerInput, "timeout") ||
		strings.Contains(lowerInput, "connection") ||
		strings.Contains(lowerInput, "device") ||
		strings.Contains(lowerInput, "communication") ||
		strings.Contains(lowerInput, "system") ||
		strings.Contains(lowerInput, "software") ||
		strings.Contains(lowerInput, "hardware") ||
		strings.Contains(lowerInput, "issue") ||
		strings.Contains(lowerInput, "failure") ||
		strings.Contains(lowerInput, "malfunction")

	// For non-technical questions, use more selective context
	if !isTechnicalQuestion {
		// Search through text files for relevant content (more selective)
		for filename, content := range b.knowledgeDB.GetTextFiles() {
			// Only include if highly relevant to user input
			relevantKeywords := 0
			inputWords := strings.Fields(lowerInput)

			for _, word := range inputWords {
				if len(word) > 3 && strings.Contains(strings.ToLower(content), word) {
					relevantKeywords++
				}
			}

			// Only include if at least 2 keywords match
			if relevantKeywords >= 2 {
				hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
				formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
				context.WriteString(fmt.Sprintf("From %s:\n", formattedPath))
				sources = append(sources, formattedPath)
				if len(content) > 400 {
					context.WriteString(content[:400] + "...\n\n")
				} else {
					context.WriteString(content + "\n\n")
				}
			}
		}

		// If still no relevant context, provide a general response
		if context.Len() == 0 {
			return fmt.Sprintf("I'm BeanBot, specifically designed for engineering support. Your question '%s' seems to be outside my technical expertise. I can help with engineering errors, system issues, device problems, and technical troubleshooting.", userInput), []string{}
		}

		return context.String(), sources
	}

	// For technical questions, use comprehensive search with priority on documentation
	// PRIORITY 1: Search HTML documentation files first (most comprehensive documentation)
	supportDocsFound := false
	for filename, content := range b.knowledgeDB.GetTextFiles() {
		// Prioritize HTML files from documentation
		if strings.Contains(strings.ToLower(filename), ".html") {
			if b.knowledgeDB.IsRelevantContent(lowerInput, content) {
				context.WriteString(fmt.Sprintf("From Engineering Documentation (%s):\n", filename))
				sources = append(sources, "Engineering Documentation: "+filename)
				if len(content) > 500 {
					context.WriteString(content[:500] + "...\n\n")
				} else {
					context.WriteString(content + "\n\n")
				}
				supportDocsFound = true
			}
		}
	}

	// PRIORITY 2: Search for relevant error codes
	for _, errorCode := range b.knowledgeDB.GetData().ErrorCodes {
		if strings.Contains(lowerInput, strings.ToLower(errorCode.Code)) ||
			strings.Contains(lowerInput, strings.ToLower(errorCode.Description)) ||
			b.knowledgeDB.ContainsAnyKeyword(lowerInput, errorCode.RelatedComponents) {

			context.WriteString(fmt.Sprintf("Error Code %s: %s\n", errorCode.Code, errorCode.Description))
			sources = append(sources, "Error Code: "+errorCode.Code)
			context.WriteString("Troubleshooting Steps:\n")
			for i, step := range errorCode.TroubleshootingSteps {
				context.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
			}
			context.WriteString("\n")
		}
	}

	// PRIORITY 3: Search for relevant common issues
	for _, issue := range b.knowledgeDB.GetData().CommonIssues {
		if strings.Contains(lowerInput, strings.ToLower(issue.Issue)) ||
			b.knowledgeDB.ContainsAnyKeyword(lowerInput, issue.Symptoms) {

			context.WriteString(fmt.Sprintf("Common Issue: %s\n", issue.Issue))
			sources = append(sources, "Common Issue: "+issue.Issue)
			context.WriteString("Solutions:\n")
			for i, solution := range issue.Solutions {
				context.WriteString(fmt.Sprintf("%d. %s\n", i+1, solution))
			}
			context.WriteString("\n")
		}
	}

	// PRIORITY 4: Search through other text files (non-HTML) for relevant content
	for filename, content := range b.knowledgeDB.GetTextFiles() {
		// Skip HTML files as they were already processed in priority 1
		if !strings.Contains(strings.ToLower(filename), ".html") {
			if b.knowledgeDB.IsRelevantContent(lowerInput, content) {
				hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
				formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
				context.WriteString(fmt.Sprintf("From %s:\n", formattedPath))
				sources = append(sources, formattedPath)
				if len(content) > 400 {
					context.WriteString(content[:400] + "...\n\n")
				} else {
					context.WriteString(content + "\n\n")
				}
			}
		}
	}

	// Search through PDF content for relevant information (skip if no readable text)
	for filename, content := range b.knowledgeDB.GetPDFContents() {
		// Skip if content looks like PDF metadata rather than text
		if strings.Contains(content, "<<") && strings.Contains(content, ">>") {
			continue // Skip PDF metadata
		}

		if b.knowledgeDB.IsRelevantContent(lowerInput, content) {
			hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
			formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
			context.WriteString(fmt.Sprintf("From %s:\n", formattedPath))
			sources = append(sources, "PDF: "+formattedPath)
			// For large PDFs like TLM, try to find the most relevant section
			if len(content) > 1000 {
				relevantSection := b.findMostRelevantSection(content, lowerInput, 800)
				context.WriteString(relevantSection + "...\n\n")
			} else if len(content) > 600 {
				context.WriteString(content[:600] + "...\n\n")
			} else {
				context.WriteString(content + "\n\n")
			}
		}
	}

	// Search through Word document content for relevant information
	for filename, content := range b.knowledgeDB.GetWordContents() {
		// Skip if content indicates an error or no content
		if strings.Contains(content, "Failed to extract") || strings.Contains(content, "not supported") {
			continue
		}

		if b.knowledgeDB.IsRelevantContent(lowerInput, content) {
			hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
			formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
			context.WriteString(fmt.Sprintf("From Word Document (%s):\n", formattedPath))
			sources = append(sources, "Word Document: "+formattedPath)
			// For large Word documents, try to find the most relevant section
			if len(content) > 1000 {
				relevantSection := b.findMostRelevantSection(content, lowerInput, 800)
				context.WriteString(relevantSection + "...\n\n")
			} else if len(content) > 600 {
				context.WriteString(content[:600] + "...\n\n")
			} else {
				context.WriteString(content + "\n\n")
			}
		}
	}

	// Search through image content for relevant information
	for filename, content := range b.knowledgeDB.GetImageContents() {
		// Skip if content indicates an error
		if strings.Contains(content, "Failed to process") {
			continue
		}

		if b.knowledgeDB.IsRelevantContent(lowerInput, content) {
			hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
			formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
			context.WriteString(fmt.Sprintf("From Image (%s):\n", formattedPath))
			sources = append(sources, "Image: "+formattedPath)
			context.WriteString(content + "\n\n")
		}
	}

	// If no specific context found, include some general troubleshooting content
	if context.Len() == 0 {
		context.WriteString("General Engineering Knowledge:\n\n")

		// If no supportDocs were found, try to include relevant HTML documentation
		if !supportDocsFound {
			// Include relevant HTML files even if not perfectly matched
			htmlCount := 0
			for filename, content := range b.knowledgeDB.GetTextFiles() {
				if strings.Contains(strings.ToLower(filename), ".html") && htmlCount < 2 {
					hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
					formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
					context.WriteString(fmt.Sprintf("From Engineering Documentation (%s):\n", formattedPath))
					sources = append(sources, "Engineering Documentation (General): "+formattedPath)
					if len(content) > 300 {
						context.WriteString(content[:300] + "...\n\n")
					} else {
						context.WriteString(content + "\n\n")
					}
					htmlCount++
				}
			}
		}

		// Include all error codes as general reference
		for _, errorCode := range b.knowledgeDB.GetData().ErrorCodes {
			context.WriteString(fmt.Sprintf("Error Code %s: %s\n", errorCode.Code, errorCode.Description))
			sources = append(sources, "Error Code Reference: "+errorCode.Code)
		}
		context.WriteString("\n")

		// Include first non-HTML text file as general reference
		for filename, content := range b.knowledgeDB.GetTextFiles() {
			if !strings.Contains(strings.ToLower(filename), ".html") {
				hierarchicalPath := b.knowledgeDB.GetFilePaths()[filename]
				formattedPath := b.formatHierarchicalReference(hierarchicalPath, filename)
				context.WriteString(fmt.Sprintf("From %s:\n", formattedPath))
				sources = append(sources, "General Reference: "+formattedPath)
				if len(content) > 300 {
					context.WriteString(content[:300] + "...\n\n")
				} else {
					context.WriteString(content + "\n\n")
				}
				break // Just include first non-HTML file for general context
			}
		}
	}

	// Limit total context size - increased limit since we have more comprehensive docs and longer responses
	result := context.String()
	if len(result) > 1500 {
		result = result[:1500] + "\n[Context truncated to prevent timeout...]"
	}

	return result, sources
}

// createEngineeringPrompt creates the prompt for Ollama
func (b *BeanBot) createEngineeringPrompt(userInput, context string) string {
	// For technical questions, use the standard engineering support format
	prompt := fmt.Sprintf(`You are BeanBot, an engineering support assistant. Analyze the user's issue and provide structured engineering guidance based on the provided knowledge base.

User Issue: %s

Knowledge Base:
%s

Provide structured engineering response:

1. PROBLEM ANALYSIS: [Identify the core issue: What is failing? What symptoms are described? What system/component is affected?]

2. SOLUTION STEPS:
   - Step 1: [First diagnostic/corrective action]
   - Step 2: [Next action based on knowledge base]
   - Step 3: [Additional verification/fix step]

3. IF PROBLEM PERSISTS: [Advanced troubleshooting or escalation steps]

Important: Base your response on the knowledge base provided. If the knowledge base contains relevant information, reference it in your solution. Analyze the user's description carefully and provide specific, actionable engineering guidance.`, userInput, context)

	return prompt
} // findMostRelevantSection finds the most relevant section of a large text for the given input
func (b *BeanBot) findMostRelevantSection(content, userInput string, maxLength int) string {
	lowerInput := strings.ToLower(userInput)

	// Split input into keywords
	keywords := strings.Fields(lowerInput)

	// Split content into sentences or paragraphs
	paragraphs := strings.Split(content, "\n\n")
	if len(paragraphs) < 3 {
		// If no clear paragraphs, split by sentences
		paragraphs = strings.Split(content, ". ")
	}

	bestScore := 0
	bestSection := ""
	currentSection := strings.Builder{}

	// Score each section based on keyword matches
	for i := range paragraphs {
		currentSection.Reset()

		// Build a section of up to 3 paragraphs
		sectionEnd := i + 3
		if sectionEnd > len(paragraphs) {
			sectionEnd = len(paragraphs)
		}

		for j := i; j < sectionEnd; j++ {
			currentSection.WriteString(paragraphs[j])
			if j < sectionEnd-1 {
				currentSection.WriteString("\n\n")
			}
		}

		section := currentSection.String()
		if len(section) > maxLength {
			continue // Skip sections that are too long
		}

		// Score this section
		score := 0
		lowerSection := strings.ToLower(section)
		for _, keyword := range keywords {
			if len(keyword) > 2 && strings.Contains(lowerSection, keyword) {
				score += strings.Count(lowerSection, keyword)
			}
		}

		if score > bestScore {
			bestScore = score
			bestSection = section
		}
	}

	// If no good section found, return the beginning
	if bestSection == "" && len(content) > maxLength {
		return content[:maxLength]
	}

	return bestSection
}

// showModelSelectionDialog shows a dialog to select available models
func (b *BeanBot) showModelSelectionDialog() {
	b.debugLog("Opening model selection dialog")

	// Check if Ollama is available
	if !b.ollamaClient.TestConnection() {
		b.debugLog("Ollama is not available for model selection")
		dialog.ShowInformation("Ollama Offline", "Ollama is not available. Please start Ollama to use AI models.", b.window)
		return
	}

	b.debugLog("Getting available models from Ollama")
	// Get available models
	models, err := b.ollamaClient.GetAvailableModels()
	if err != nil {
		b.debugLog("Failed to get available models: %v", err)
		dialog.ShowError(fmt.Errorf("failed to get available models: %w", err), b.window)
		return
	}

	b.debugLog("Found %d available models: %v", len(models), models)
	if len(models) == 0 {
		dialog.ShowInformation("No Models", "No models are installed. Please install a model using:\n\nollama pull llama3.2:1b", b.window)
		return
	}

	// Get current model
	currentModel := b.ollamaClient.GetCurrentModel()
	b.debugLog("Current model: %s", currentModel)

	// Create selection list
	list := widget.NewList(
		func() int { return len(models) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Model")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			model := models[id]

			if model == currentModel {
				label.SetText(fmt.Sprintf("✓ %s (current)", model))
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.SetText(fmt.Sprintf("  %s", model))
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedModel := models[id]
		b.debugLog("Model selected: %s (was: %s)", selectedModel, currentModel)
		if selectedModel != currentModel {
			// Update the model
			b.ollamaClient.SetModel(selectedModel)
			b.debugLog("Model changed to: %s", selectedModel)
			// Update the status label
			b.statusLabel.SetText(fmt.Sprintf("🤖 BeanBot AI - %s ✅ ready to help! (click to change)", selectedModel))
		}
	}

	// Create dialog with larger size
	scrollContainer := container.NewScroll(list)
	scrollContainer.Resize(fyne.NewSize(500, 400)) // Set explicit size for better visibility

	dialogContent := container.NewBorder(
		widget.NewLabel("Available Models:"),
		nil, nil, nil,
		scrollContainer,
	)
	dialogContent.Resize(fyne.NewSize(520, 450)) // Set size for the entire dialog content

	dialog.ShowCustom("Select AI Model", "Close", dialogContent, b.window)
}

// EnableDebugMode enables debug logging
func (b *BeanBot) EnableDebugMode() {
	b.debugMode = true
	log.Println("[DEBUG] Debug mode enabled")
}

// debugLog logs debug information if debug mode is enabled
func (b *BeanBot) debugLog(format string, args ...interface{}) {
	if b.debugMode {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// formatHierarchicalReference formats document references with hierarchical folder paths
func (b *BeanBot) formatHierarchicalReference(filepath, filename string) string {
	if filepath == "" {
		return filename
	}

	// Keep the full filename with extension
	baseName := filename

	// Get the directory parts
	dir := filepath
	if strings.HasSuffix(dir, filename) {
		dir = strings.TrimSuffix(dir, filename)
		dir = strings.TrimSuffix(dir, "/")
		dir = strings.TrimSuffix(dir, "\\")
	}

	// Remove testData prefix
	dir = strings.TrimPrefix(dir, "testData/")
	dir = strings.TrimPrefix(dir, "testData\\")
	dir = strings.TrimPrefix(dir, "testData")

	if dir == "" || dir == "." {
		return baseName
	}

	// Split directory path and format hierarchically
	parts := strings.Split(strings.ReplaceAll(dir, "\\", "/"), "/")
	var pathParts []string
	for _, part := range parts {
		if part != "" && part != "." && part != "testData" {
			pathParts = append(pathParts, part)
		}
	}

	if len(pathParts) == 0 {
		return baseName
	}

	// Format as parent/reference (no brackets, skip testData grandparent)
	if len(pathParts) == 1 {
		return fmt.Sprintf("%s/%s", pathParts[0], baseName)
	} else if len(pathParts) >= 2 {
		// Show last folder and file
		return fmt.Sprintf("%s/%s", pathParts[len(pathParts)-1], baseName)
	}

	return baseName
}
