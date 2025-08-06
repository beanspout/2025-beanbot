# BeanBot - Engineering Support Assistant

BeanBot is a Go-based desktop application that provides AI-powered engineering support using Ollama models. It features a clean GUI interface for troubleshooting technical issues with integrated knowledge base search and file upload capabilities.

## 🚀 Quick Start

### Prerequisites
- Go 1.24+ 
- [Ollama](https://ollama.ai/) installed and running
- Recommended model: `ollama pull llama3.2:1b`

### Building & Running
```bash
# Clone the repository
git clone https://github.com/beanspout/2025-beanbot.git
cd 2025-beanbot

# Build the application (Windows)
go build -o lsie-beanbot.exe

# Build the application (Linux/macOS)
go build -o lsie-beanbot

# Run the application (Windows)
./lsie-beanbot.exe

# Run the application (Linux/macOS)
./lsie-beanbot

# On Windows, the executable will be 'lsie-beanbot.exe':
# lsie-beanbot.exe
```

## 📁 Codebase Architecture

### Core Structure Overview
```
├── main.go                 # Application entry point
├── internal/               # Private application code
│   ├── ui/                 # User interface layer
│   ├── knowledge/          # Knowledge database management
│   ├── ollama/             # AI model integration
│   └── models/             # Data structures
├── pkg/                    # File processing utilities
├── testData/               # Knowledge base content
└── output examples/        # Sample outputs
```

## 🔧 Key Components & Where to Find Them

### 🎨 User Interface (`internal/ui/`)

**`app.go`** - Main UI controller (917 lines)
- **Function: `SetupUI()`** (Line ~38) - Initializes the main window layout
- **Function: `createFooter()`** (Line ~58) - Status bar and model selection dropdown
- **Function: `createMainContent()`** (Line ~169) - Chat interface with input/response areas
- **Function: `handleEngineeringRequest()`** (Line ~240) - Core request processing logic
- **Function: `buildEngineeringContext()`** (Line ~440) - Knowledge base search and context building
- **Function: `createEngineeringPrompt()`** (Line ~750) - AI prompt generation for structured responses

**`file_dialog.go`** - Windows file upload dialogs (199 lines)
- **Function: `ShowFileDialog()`** - Native Windows file picker integration
- Uses Windows API calls for seamless file selection

### 🧠 Knowledge Management (`internal/knowledge/`)

**`database.go`** - Knowledge base engine (833 lines)
- **Function: `NewKnowledgeDatabase()`** (Line ~30) - Initializes and loads all knowledge sources
- **Function: `ProcessUserUpload()`** (Line ~100+) - Handles user file uploads and processing
- **Function: `IsRelevantContent()`** (Line ~300+) - Smart content relevance detection
- **File Processing Methods:**
  - `processTextFiles()` - Handles .txt, .html, .json files
  - `processPDFFiles()` - Extracts text from PDF documents
  - `processWordFiles()` - Extracts content from .docx files
  - `processImageFiles()` - OCR and image content analysis

### 🤖 AI Integration (`internal/ollama/`)

**`client.go`** - Ollama API client (428 lines)
- **Function: `NewClient()`** (Line ~18) - Creates configured HTTP client with 2-minute timeout
- **Function: `TestConnection()`** (Line ~29) - Validates Ollama server connectivity
- **Function: `FindAvailableModel()`** (Line ~37) - Auto-detects best available model
- **Function: `GenerateResponse()`** (Line ~100+) - Sends prompts and handles AI responses
- **Function: `GetAvailableModels()`** (Line ~200+) - Lists all installed Ollama models

### 📊 Data Models (`internal/models/`)

**`types.go`** - Core data structures (45 lines)
- **`TroubleshootingData`** - Main knowledge base structure
- **`ErrorCode`** - Structured error code definitions with troubleshooting steps
- **`CommonIssue`** - Frequent problems and their solutions
- **`OllamaRequest/Response`** - API communication structures

### 🔄 File Processors (`pkg/`)

**Specialized file handlers for different formats:**
- **`pdf_processor.go`** - PDF text extraction using github.com/ledongthuc/pdf
- **`windows_image_processor.go`** - OCR processing using Windows APIs
- **`drawio_processor.go`** - Draw.io diagram processing
- **`processors/`** - Additional format-specific processors

## 🎯 Key Features & Implementation

### 🔍 Smart Context Building
**Location:** `internal/ui/app.go` → `buildEngineeringContext()`
- **Priority System:** User uploads → Error codes → HTML docs → Text files → PDFs
- **Relevance Detection:** Keyword matching with technical content scoring
- **Content Limiting:** Prevents context overflow with intelligent truncation

### 💬 Structured AI Responses  
**Location:** `internal/ui/app.go` → `createEngineeringPrompt()`
- **Response Format:** Problem Analysis → Solution Steps → Advanced Troubleshooting
- **Source Attribution:** Always includes referenced knowledge base sources
- **Markdown Rendering:** Rich text formatting with bold headers and bullet lists

### 📤 File Upload System
**Location:** `internal/ui/file_dialog.go` + `internal/knowledge/database.go`
- **Native Windows Dialog:** Uses Windows API for seamless file selection
- **Multi-format Support:** PDF, Word, images, text files, Draw.io diagrams
- **Session Management:** User uploads are temporary and cleared with "Clear" button

### 🎛️ Model Management
**Location:** `internal/ui/app.go` → `createFooter()` + `internal/ollama/client.go`
- **Auto-detection:** Scans for available Ollama models on startup
- **Dynamic Switching:** Runtime model switching with UI updates
- **Fallback Logic:** Tries multiple models if preferred isn't available

## 🚀 Development Guide

### Adding New File Format Support
1. Create processor in `pkg/processors/[format]_processor.go`
2. Add processing logic to `internal/knowledge/database.go`
3. Update file filter in `internal/ui/file_dialog.go`

### Modifying AI Response Format
1. Edit prompt template in `createEngineeringPrompt()` (`internal/ui/app.go`)
2. Adjust markdown parsing in `handleEngineeringRequest()`
3. Update source reference formatting in `buildEngineeringContext()`

### Extending Knowledge Base
1. Add new data structures to `internal/models/types.go`
2. Update loading logic in `internal/knowledge/database.go`
3. Modify context building in `buildEngineeringContext()`

## 📋 Dependencies

### Core Framework
- **Fyne v2.4.5** - Cross-platform GUI framework
- **Go 1.24+** - Backend language with modern features

### File Processing
- **github.com/ledongthuc/pdf** - PDF text extraction
- **github.com/nguyenthenguyen/docx** - Word document processing
- **github.com/go-ole/go-ole** - Windows COM/OLE integration for image OCR

### AI Integration
- **Ollama** - Local AI model serving (external dependency)
- **HTTP Client** - Standard library for API communication

## 🎨 UI Architecture

### Layout Pattern
- **Border Layout:** Fixed footer + scrollable content area
- **Chat Interface:** Input at bottom, responses above (familiar messaging pattern)
- **Responsive Design:** Auto-wrapping text and dynamic sizing

### State Management
- **Session-based:** User uploads cleared on "Clear" button
- **Real-time Updates:** Status bar reflects current model and connection state
- **Progressive Loading:** Async model detection with UI feedback

## 🔧 Configuration

### Default Settings (`main.go`)
- **Window Size:** 450x700 (optimized for chat interface)
- **Default Model:** llama3.2:1b (lightweight and fast)
- **Ollama URL:** http://localhost:11434 (standard Ollama port)
- **Request Timeout:** 120 seconds (allows for larger model responses)

### Knowledge Base Location
- **Primary Data:** `testData/` directory contains all knowledge sources
- **Error Codes:** `testData/lsie_errors.json` - structured troubleshooting data
- **Documentation:** `testData/Confluence/` - HTML documentation files
- **Test Files:** `testData/` - sample text and configuration files

## 🐛 Debugging

### Debug Mode
Enable with `bot.EnableDebugMode()` in `main.go` for detailed logging:
- Model selection and switching events
- Context building and source selection
- File processing results
- Ollama API communication details

### Common Issues
- **Ollama Offline:** Check if `ollama serve` is running
- **No Models:** Install with `ollama pull llama3.2:1b`
- **File Processing Errors:** Check file permissions and format support
- **Response Timeout:** Reduce context size or use smaller model

---
