# BeanBot Development and Testing Guide

## Prerequisites

**Important**: Make sure you're in the correct project directory before running any commands. The setup script must be run from the directory containing `go.mod`, `main.go`, and the `internal/` folder.

```bash
# Navigate to the project directory first
cd c:\Users\NZ26RQ\source\repos\BeanBot

# Verify you're in the right place
dir
# You should see: go.mod, main.go, internal/, testData/, etc.
```

## Quick Start

1. **Run the setup script** (Windows):
   ```
   setup.bat
   ```
   This will:
   - Check for Ollama installation
   - Download gemma3:1b model if needed
   - Start Ollama service
   - Build and launch BeanBot

2. **Manual setup**:
   ```
   # Install Ollama (if not installed)
   # Download from https://ollama.ai
   
   # Pull the required model
   ollama pull gemma3:1b
   
   # Start Ollama
   ollama serve
   
   # Build and run BeanBot
   go build -o beanbot.exe ./main.go
   ./beanbot.exe
   ```

## Testing the Application

### 1. Basic Troubleshooting Test
- Start BeanBot
- Enter: "I'm getting error E1001 when trying to communicate with the device"
- Click "Get Troubleshooting Help"
- Expected: AI response with step-by-step communication troubleshooting

### 2. File Upload Tests

#### PDF Upload Test
- Click "Upload PDF Documentation"
- Select any PDF file (or use sample documentation)
- Expected: Success message with extracted keywords

#### Screenshot Upload Test
- Click "Upload Screenshot"
- Select an image file (.png, .jpg, .bmp)
- Expected: Analysis results with UI elements and error patterns

#### DrawIO Upload Test
- Click "Upload DrawIO Diagram"
- Select a .drawio file from testData/troubleTree/
- Expected: Diagram analysis with components and flow information

### 3. Error Code Recognition Test
Try these specific error scenarios:
- "Error E2005 temperature sensor fault"
- "E3010 power supply voltage issue"
- "Communication timeout with VICM"
- "Sensor calibration failure"

## Troubleshooting Common Issues

### Ollama Not Responding
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Restart Ollama
ollama serve

# Verify gemma3:1b model is available
ollama list

# Check if required model is available
ollama list | findstr "gemma3:1b"
```

### Model Issues
```bash
# BeanBot requires gemma3:1b model specifically
ollama pull gemma3:1b

# Verify the model downloaded correctly
ollama list | findstr "gemma3:1b"

# If model is missing, BeanBot will show an error
# Download the required model:
ollama pull gemma3:1b
```

### Build Errors
```bash
# Clean and rebuild
go clean
go mod tidy
go build

# If you get "no required module provides package" errors:
# 1. Ensure you're in the correct directory (should contain go.mod)
cd c:\path\to\BeanBot

# 2. Check the go.mod file exists and has correct module name
type go.mod

# 3. Verify internal packages exist
dir internal\

# 4. Run build with verbose output to see what's happening
go build -v ./main.go

# 5. If still failing, try rebuilding the module
go mod init github.com/beanspout/2025-beanbot
go mod tidy
go build ./main.go
```

### Missing Dependencies
```bash
# Update dependencies
go get -u
go mod tidy
```

## Development Tips

### Adding New Error Codes
Edit `testData/itest_errors.json`:
```json
{
  "code": "E4001",
  "description": "New error type",
  "category": "custom",
  "severity": "medium",
  "troubleshooting_steps": [
    "Step 1: Check this",
    "Step 2: Verify that"
  ],
  "related_components": ["Component1"],
  "documentation_reference": "manual.pdf"
}
```

### Customizing AI Prompts
Modify the `createTroubleshootingPrompt` function in `main.go` to adjust how the AI responds.

### Extending File Processors
- **PDF**: Enhance `pdf_processor.go` with actual text extraction
- **Images**: Improve `windows_image_processor.go` with OCR
- **DrawIO**: Extend `drawio_processor.go` with more diagram analysis

## Performance Optimization

### For Lab Computers
- Use local Ollama installation (no internet required)
- Enable Windows API optimizations
- Minimize memory usage with smaller model variants

### For Workstations
- Consider larger models for better accuracy
- Enable additional image processing features
- Implement caching for frequently accessed documents

## Integration with iTest

### Custom Knowledge Base
1. Export iTest error logs to JSON format
2. Add custom troubleshooting procedures
3. Include component manuals and documentation
4. Update knowledge base loading in `main.go`

### Automated Screenshot Analysis
1. Configure screenshot capture hotkeys
2. Implement automatic error detection
3. Set up watch folders for new screenshots
4. Enable batch processing of multiple images

## Security Considerations

### Lab Environment
- Restrict file access to authorized directories
- Implement user authentication if needed
- Log all troubleshooting sessions
- Sanitize file uploads

### Data Privacy
- Keep all processing local (no cloud APIs)
- Encrypt sensitive documentation
- Implement secure file handling
- Clear temporary files on exit

## Deployment

### Single Executable Distribution
```bash
# Build for Windows
GOOS=windows GOARCH=amd64 go build -o beanbot.exe

# Create installation package
# Include: beanbot.exe, testData/, config.json, README.md
```

### Lab Computer Installation
1. Copy BeanBot folder to shared location
2. Install Ollama on each machine
3. Configure network access if needed
4. Create desktop shortcuts for users

## Monitoring and Maintenance

### Log Analysis
- Monitor `beanbot.log` for errors
- Track usage patterns
- Identify common issues
- Update knowledge base accordingly

### Model Updates
```bash
# Update to newer model versions
ollama pull gemma3:1b
# BeanBot is configured to use gemma3:1b specifically
```

### Knowledge Base Maintenance
- Regular updates to error codes
- Add new troubleshooting procedures
- Include latest documentation
- Remove outdated information
