package ui

import (
	"fmt"
	"log"
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
	app          fyne.App
	window       fyne.Window
	knowledgeDB  *knowledge.KnowledgeDatabase
	ollamaClient *ollama.Client
	submitBtn    *widget.Button
	statusLabel  *widget.Label // Add reference to status label for updates
	debugMode    bool          // Debug mode flag
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
		b.debugLog("Testing Ollama connection to %s", b.ollamaClient.TestConnection())
		if b.ollamaClient.TestConnection() {
			b.debugLog("Ollama connection successful, searching for available models")
			// Try to get available models
			available, model := b.ollamaClient.FindAvailableModel()
			if available {
				b.debugLog("Found available model: %s", model)
				status.SetText(fmt.Sprintf("🤖 BeanBot AI - %s ✅ ready to help! (click to change)", model))
				// Update the client to use the found model
				b.ollamaClient.SetModel(model)
				b.debugLog("Set active model to: %s", model)
			} else {
				b.debugLog("No models found available")
				status.SetText("🤖 BeanBot AI ❌ models missing")
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
	inputEntry.SetPlaceHolder("Describe your LSIE issue... (Example: I'm getting error E1001 when trying to communicate with the device)")
	inputEntry.Wrapping = fyne.TextWrapWord   // Enable word wrapping for input too
	inputEntry.Resize(fyne.NewSize(800, 100)) // Set a reasonable height for input

	// Create response area using RichText with Border Layout Pattern (no white space)
	responseText := widget.NewRichTextFromMarkdown("\n\n\n\n## 🤖 Hi there! \n\n### What can I help you troubleshoot today? 💭")
	responseText.Wrapping = fyne.TextWrapWord

	// Create submit button with handler for RichText
	submitBtn := widget.NewButton("Ask", func() {
		b.handleTroubleshootingRequest(inputEntry.Text, responseText)
	})
	submitBtn.Importance = widget.HighImportance

	// Create clear button to reset everything
	clearBtn := widget.NewButton("Clear", func() {
		// Clear the input field
		inputEntry.SetText("")
		// Reset response area to welcome message
		responseText.ParseMarkdown("\n\n\n\n## 🤖 Hi there! \n\n### What can I help you troubleshoot today? 💭")
	})

	// Store reference to button for progress handling
	b.submitBtn = submitBtn

	// Fixed content for bottom section (input area) - clean chat-style layout with equal-sized buttons
	buttonContainer := container.NewGridWithColumns(2, submitBtn, clearBtn)
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
	mainContainer := container.NewBorder(
		nil,                                  // Top - not needed (no header)
		bottomSection,                        // Bottom - fixed size input area (chat-style)
		nil,                                  // Left - not needed
		nil,                                  // Right - not needed
		container.NewScroll(centeredContent), // Center - scrollable centered RichText (takes remaining space)
	)

	return mainContainer
}

// handleTroubleshootingRequest handles the troubleshooting request
func (b *BeanBot) handleTroubleshootingRequest(userInput string, responseEntry *widget.RichText) {
	if strings.TrimSpace(userInput) == "" {
		dialog.ShowError(fmt.Errorf("please describe your issue"), b.window)
		return
	}

	b.debugLog("Handling troubleshooting request: %s", userInput)
	b.debugLog("Current model: %s", b.ollamaClient.GetCurrentModel())

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

		b.debugLog("Building troubleshooting context...")
		// Build context from knowledge database
		context := b.buildTroubleshootingContext(userInput)
		b.debugLog("Context length: %d characters", len(context))

		// Create prompt for Ollama
		prompt := b.createTroubleshootingPrompt(userInput, context)
		b.debugLog("Prompt length: %d characters", len(prompt))

		// Check if this is a direct response (not a prompt for Ollama)
		if strings.Contains(context, "outside my technical troubleshooting expertise") {
			b.debugLog("Using direct response (outside expertise)")
			responseEntry.ParseMarkdown(context)
			return
		}

		b.debugLog("Sending request to Ollama with model: %s", b.ollamaClient.GetCurrentModel())
		// Get response from Ollama
		response, err := b.ollamaClient.GenerateResponse(prompt)
		if err != nil {
			b.debugLog("Error getting AI response: %v", err)
			log.Printf("Error getting AI response: %v", err)
			responseEntry.ParseMarkdown(fmt.Sprintf("Error getting AI response: %v", err))
			return
		}

		b.debugLog("Received response from Ollama, length: %d characters", len(response))
		// Display response in the same window
		responseEntry.ParseMarkdown(response)
	}()
}

// buildTroubleshootingContext builds context from the knowledge database
func (b *BeanBot) buildTroubleshootingContext(userInput string) string {
	var context strings.Builder
	lowerInput := strings.ToLower(userInput)

	// Check if this is a technical troubleshooting question vs general question
	isTechnicalQuestion := strings.Contains(lowerInput, "error") ||
		strings.Contains(lowerInput, "problem") ||
		strings.Contains(lowerInput, "troubleshoot") ||
		strings.Contains(lowerInput, "timeout") ||
		strings.Contains(lowerInput, "connection") ||
		strings.Contains(lowerInput, "device") ||
		strings.Contains(lowerInput, "communication") ||
		strings.Contains(lowerInput, "e1001") ||
		strings.Contains(lowerInput, "e2005") ||
		strings.Contains(lowerInput, "e3010")

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
				context.WriteString(fmt.Sprintf("From %s:\n", filename))
				if len(content) > 400 {
					context.WriteString(content[:400] + "...\n\n")
				} else {
					context.WriteString(content + "\n\n")
				}
			}
		}

		// If still no relevant context, provide a general response
		if context.Len() == 0 {
			return fmt.Sprintf("I'm BeanBot, specifically designed for LSIE troubleshooting. Your question '%s' seems to be outside my technical troubleshooting expertise. I can help with LSIE errors, communication issues, device problems, and technical troubleshooting.", userInput)
		}

		return context.String()
	}

	// For technical questions, use comprehensive search with priority on LSIE_SupportDocs
	// PRIORITY 1: Search LSIE_SupportDocs HTML files first (most comprehensive documentation)
	supportDocsFound := false
	for filename, content := range b.knowledgeDB.GetTextFiles() {
		// Prioritize HTML files from LSIE_SupportDocs
		if strings.Contains(strings.ToLower(filename), ".html") {
			if b.knowledgeDB.IsRelevantContent(lowerInput, content) {
				context.WriteString(fmt.Sprintf("From LSIE Documentation (%s):\n", filename))
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
				context.WriteString(fmt.Sprintf("From %s:\n", filename))
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
			context.WriteString(fmt.Sprintf("From %s:\n", filename))
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

	// If no specific context found, include some general troubleshooting content
	if context.Len() == 0 {
		context.WriteString("General iTest Troubleshooting Knowledge:\n\n")

		// If no supportDocs were found, try to include relevant HTML documentation
		if !supportDocsFound {
			// Include relevant HTML files even if not perfectly matched
			htmlCount := 0
			for filename, content := range b.knowledgeDB.GetTextFiles() {
				if strings.Contains(strings.ToLower(filename), ".html") && htmlCount < 2 {
					context.WriteString(fmt.Sprintf("From iTest Documentation (%s):\n", filename))
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
		}
		context.WriteString("\n")

		// Include first non-HTML text file as general reference
		for filename, content := range b.knowledgeDB.GetTextFiles() {
			if !strings.Contains(strings.ToLower(filename), ".html") {
				context.WriteString(fmt.Sprintf("From %s:\n", filename))
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

	return result
}

// createTroubleshootingPrompt creates the prompt for Ollama
func (b *BeanBot) createTroubleshootingPrompt(userInput, context string) string {
	// For technical questions, use the standard troubleshooting format
	prompt := fmt.Sprintf(`You are BeanBot, an iTest troubleshooting assistant. Analyze the user's issue and provide structured troubleshooting guidance.

User Issue: %s

Knowledge Base:
%s

Provide structured troubleshooting response:

1. PROBLEM ANALYSIS: [Identify the core issue: What is failing? What symptoms are described? What system/component is affected?]

2. SOLUTION STEPS:
   - Step 1: [First diagnostic/corrective action]
   - Step 2: [Next action based on knowledge base]
   - Step 3: [Additional verification/fix step]

3. IF PROBLEM PERSISTS: [Advanced troubleshooting or escalation steps]

Analyze the user's description carefully and provide specific, actionable guidance based on the knowledge base.`, userInput, context)

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
	// Check if Ollama is available
	if !b.ollamaClient.TestConnection() {
		dialog.ShowInformation("Ollama Offline", "Ollama is not available. Please start Ollama to use AI models.", b.window)
		return
	}

	// Get available models
	models, err := b.ollamaClient.GetAvailableModels()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to get available models: %w", err), b.window)
		return
	}

	if len(models) == 0 {
		dialog.ShowInformation("No Models", "No models are installed. Please install a model using 'ollama pull <model>'", b.window)
		return
	}

	// Get current model
	currentModel := b.ollamaClient.GetCurrentModel()

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
		if selectedModel != currentModel {
			// Update the model
			b.ollamaClient.SetModel(selectedModel)
			// Update the status label
			b.statusLabel.SetText(fmt.Sprintf("🤖 BeanBot AI - %s ✅ ready to help! (click to change)", selectedModel))
		}
	}

	// Create dialog
	dialog.ShowCustom("Select AI Model", "Close", container.NewBorder(
		widget.NewLabel("Available Models:"),
		nil, nil, nil,
		container.NewScroll(list),
	), b.window)
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
