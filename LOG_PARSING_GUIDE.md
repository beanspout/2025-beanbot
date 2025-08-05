# Log File Parsing Support in BeanBot

## Overview
BeanBot now includes comprehensive log file parsing capabilities to help with troubleshooting and error analysis. When you upload log files, BeanBot automatically detects, parses, and analyzes them to extract meaningful information.

## Supported Log File Formats

### File Extensions
- `.log` - Standard log files
- `.logs` - Log collections
- `.txt` files containing log-like content (auto-detected)

### Auto-Detection
BeanBot can automatically detect log files even when they have `.txt` extensions by looking for:
- Log level indicators: [DEBUG], [INFO], [WARN], [ERROR], [FATAL]
- Timestamp patterns
- Common log keywords: exception, stack trace, timeout, retry, etc.

## Log Analysis Features

### 1. **Structured Summary**
- Total line count
- Error count
- Warning count  
- Exception count
- Important events count
- Time range analysis

### 2. **Categorized Extraction**
- **Errors**: Lines containing "error", "failed", "fatal", "[error]"
- **Warnings**: Lines with "warn", "warning", "[warn]"
- **Exceptions**: Stack traces, exceptions, throwables
- **Important Events**: Started/stopped, connected/disconnected, timeouts, retries, configuration changes

### 3. **Smart Context Prioritization**
- User-uploaded log files get PRIORITY 0 (highest priority)
- Liberal inclusion criteria for log content
- Enhanced content space (800 characters vs 400 for regular files)

## Usage Instructions

### 1. **Upload Log Files**
1. Click "Upload Files" button
2. Select your log files (.log, .logs, or .txt files)
3. BeanBot will automatically parse and analyze them

### 2. **Ask Questions**
After uploading log files, you can ask questions like:
- "What errors are in my log file?"
- "Why is my application crashing?"
- "What happened before the timeout?"
- "Analyze the connection issues"

### 3. **Get Intelligent Responses**
BeanBot will prioritize information from your uploaded logs and provide:
- Error analysis
- Timeline reconstruction
- Troubleshooting recommendations
- Context-aware solutions

## Example Log Analysis Output

```
=== LOG FILE ANALYSIS: application.log ===

**LOG SUMMARY:**
- Total lines: 15
- Errors found: 2
- Warnings found: 2
- Exceptions found: 1
- Important events: 8

**TIME RANGE:**
First entry: 2025-08-05 14:30:25 [INFO] Application started successfully
Last entry: 2025-08-05 14:32:02 [INFO] Cleanup completed, application stopped

**ERRORS FOUND:**
1. 2025-08-05 14:30:32 [ERROR] Failed to connect to remote server: Connection refused
2. 2025-08-05 14:31:16 [ERROR] Validation failed: Invalid input parameter 'username'

**WARNINGS FOUND:**
1. 2025-08-05 14:30:30 [WARN] Connection timeout detected, retrying...
2. 2025-08-05 14:31:17 [WARN] Rate limit exceeded for IP 192.168.1.100

**EXCEPTIONS/STACK TRACES:**
1. Exception in thread "main" java.lang.OutOfMemoryError: Java heap space
```

## Benefits

### 1. **Faster Troubleshooting**
- Automatically extracts relevant errors and warnings
- Provides timeline context
- Identifies patterns and root causes

### 2. **Comprehensive Analysis**
- Processes large log files efficiently (handles up to 1000 lines)
- Categorizes different types of log entries
- Maintains chronological context

### 3. **Smart Integration**
- Seamlessly integrates with existing knowledge base
- Prioritizes user-uploaded content
- Provides context-aware recommendations

## Technical Implementation

### Auto-Detection Algorithm
```go
// Checks for log indicators:
- Log level keywords: [DEBUG], [INFO], [WARN], [ERROR], [FATAL]
- Timestamp patterns
- Common log terms: exception, stack trace, timeout, retry
- System events: started, stopped, connected, failed
```

### Parsing Logic
```go
// Extracts and categorizes:
- Errors and failures
- Warnings and alerts  
- Exceptions and stack traces
- Important system events
- Timeline information
```

This feature significantly enhances BeanBot's ability to provide intelligent troubleshooting assistance by understanding the actual logs and errors from your systems.
