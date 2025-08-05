# BeanBot - iTest Troubleshooting Assistant

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne-GUI-blue?style=for-the-badge)
![AI](https://img.shields.io/badge/AI-Powered-green?style=for-the-badge)

BeanBot is an intelligent troubleshooting assistant specifically designed for iTest systems. It combines comprehensive technical documentation with AI-powered analysis to provide expert-level troubleshooting guidance.

## 🚀 Features

### 🧠 AI-Powered Analysis
- Integration with **Ollama** using the `gemma3:1b` model for intelligent responses
- Smart context building that prioritizes official iTest documentation
- Fallback responses when AI service is unavailable
- Optimized for small language models with intelligent context limiting

### 📚 Comprehensive Knowledge Base
- **216 iTest Support Documents** - Complete HTML documentation library
- **Multi-format Support** - JSON, TXT, PDF, DrawIO, and HTML files
- **Priority-based Search** - Official documentation prioritized over basic error codes
- **PDF Text Extraction** - Advanced parsing with encoding fixes
- **Real-time Content Processing** - Dynamic loading of troubleshooting data

### 🖥️ Modern User Interface
- **Cross-platform GUI** built with Fyne v2.4.5
- **Resizable Panels** - Adjustable input/response sections with VSplit containers
- **Word Wrapping & Scrolling** - Responsive text display with proper formatting
- **Real-time Processing** - Live progress indicators and status updates
- **Professional Design** - Clean, intuitive interface optimized for technical users

### 🔧 Advanced Technical Capabilities
- **Professional Go Architecture** - Clean internal package structure
- **Multi-format Document Processing** - Handles diverse technical documentation
- **Smart Relevance Matching** - Context-aware content selection
- **Performance Optimization** - Efficient memory usage and fast response times
- **Comprehensive Error Handling** - Robust logging and fallback mechanisms

## 📋 Prerequisites

- **Go 1.21 or later**
- **Ollama** (optional, for AI features)
- **Windows/Linux/macOS** (cross-platform support)

## 🛠️ Installation

### Option 1: Quick Setup (Recommended)
```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/BeanBot.git
cd BeanBot

# Run the setup script (Windows)
setup.bat

# The script will:
# - Install Go dependencies
# - Set up Ollama with gemma3:1b model
# - Build the application
# - Launch BeanBot
```

### Option 2: Manual Installation
```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/BeanBot.git
cd BeanBot

# Install dependencies
go mod download

# Build the application
go build -o beanbot main.go

# Install Ollama (optional but recommended)
# Visit https://ollama.ai/ for installation instructions
ollama pull gemma3:1b

# Run BeanBot
./beanbot
```

## 🏃‍♂️ Quick Start

1. **Launch BeanBot**
   ```bash
   ./beanbot
   ```

2. **Describe Your Issue**
   - Enter your iTest problem in the input area
   - Be specific about error codes, symptoms, or components
   - Example: "Getting error E1001 when trying to communicate with the device"

3. **Get Expert Guidance**
   - BeanBot analyzes your issue against the knowledge base
   - Receives structured troubleshooting steps
   - Follows priority-based recommendations from official documentation

## 📁 Project Structure

```
BeanBot/
├── cmd/
│   └── beanbot/           # Alternative entry point
├── internal/              # Private application code
│   ├── knowledge/         # Knowledge database management
│   ├── models/           # Data structures and types
│   ├── ollama/           # AI client integration
│   └── ui/               # User interface components
├── pkg/                  # Public packages
│   └── processors/       # Document processing utilities
├── testData/             # Knowledge base content
│   ├── iTest_SupportDocs/ # 216 HTML documentation files
│   ├── troubleTree/      # Troubleshooting diagrams
│   └── *.{json,txt,pdf}  # Additional technical resources
├── main.go               # Application entry point
├── setup.bat            # Windows setup script
└── README.md            # This file
```

## 🔬 Technical Architecture

### Knowledge Database
- **Priority System**: iTest documentation → Error codes → Common issues → General content
- **Smart Caching**: Efficient memory management for large document sets
- **Relevance Scoring**: Advanced keyword matching with technical term awareness
- **Multi-format Parsing**: Specialized processors for each document type

### AI Integration
- **Ollama Client**: HTTP-based communication with local AI models
- **Context Optimization**: Intelligent truncation for small model compatibility
- **Fallback System**: Graceful degradation when AI is unavailable
- **Response Enhancement**: Structured formatting for technical guidance

### User Interface
- **Responsive Design**: Adaptive layout for different screen sizes
- **Real-time Updates**: Live progress indication and status monitoring
- **Accessibility**: Keyboard navigation and screen reader support
- **Cross-platform**: Native look and feel on all supported systems

## 🤝 Contributing

We welcome contributions to improve BeanBot! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'Add some amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Development Guidelines
- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for API changes
- Ensure cross-platform compatibility

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/YOUR_USERNAME/BeanBot/issues)
- **Documentation**: Check the `testData/iTest_SupportDocs/` directory for comprehensive iTest documentation
- **Community**: Join our discussions in [GitHub Discussions](https://github.com/YOUR_USERNAME/BeanBot/discussions)

## 🔮 Roadmap

- [ ] **Web Interface** - Browser-based access for remote troubleshooting
- [ ] **API Endpoints** - RESTful API for integration with other tools
- [ ] **Plugin System** - Extensible architecture for custom processors
- [ ] **Multiple AI Models** - Support for different language models
- [ ] **Advanced Analytics** - Usage patterns and troubleshooting effectiveness
- [ ] **Cloud Deployment** - Docker containers and cloud hosting options

## 🏆 Acknowledgments

- **iTest Documentation Team** - For comprehensive technical documentation
- **Ollama Project** - For providing accessible local AI capabilities
- **Fyne Project** - For the excellent cross-platform GUI framework
- **Go Community** - For the robust programming language and ecosystem

---

**Made with ❤️ for the iTest community**