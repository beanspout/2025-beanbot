@echo off
echo BeanBot - iTest Troubleshooting Assistant Setup
echo ================================================

:: Check if Ollama is installed
where ollama >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Ollama is not installed or not in PATH
    echo Please install Ollama from https://ollama.ai
    echo.
    echo After installation, this script will automatically download the recommended model.
    pause
    exit /b 1
)

:: Check if gemma3:1b model is available
echo Checking for gemma3:1b model...
ollama list | findstr "gemma3:1b" >nul
if %errorlevel% neq 0 (
    echo Gemma3:1b model not found. Downloading...
    echo This may take a few minutes depending on your internet connection.
    ollama pull gemma3:1b
    if %errorlevel% neq 0 (
        echo ERROR: Failed to download gemma3:1b model
        echo Please check your internet connection and try again
        echo Or manually run: ollama pull gemma3:1b
        pause
        exit /b 1
    )
) else (
    echo Found gemma3:1b model - ready for troubleshooting!
)

:: Check if Ollama service is running
echo Checking Ollama service...
curl -s http://localhost:11434/api/tags >nul 2>nul
if %errorlevel% neq 0 (
    echo Starting Ollama service...
    start /b ollama serve
    echo Waiting for Ollama to start...
    timeout /t 10 /nobreak >nul
    
    :: Verify service started
    curl -s http://localhost:11434/api/tags >nul 2>nul
    if %errorlevel% neq 0 (
        echo WARNING: Ollama service may not have started properly
        echo You may need to manually run: ollama serve
        echo Press any key to continue anyway...
        pause >nul
    )
)

:: Build the application
echo Building BeanBot...
go version >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

go build -o beanbot.exe ./main.go
if %errorlevel% neq 0 (
    echo ERROR: Failed to build BeanBot
    echo Trying to fix dependencies...
    go mod tidy
    go build -o beanbot.exe ./main.go
    if %errorlevel% neq 0 (
        echo ERROR: Build failed even after fixing dependencies
        echo Please check the error messages above
        pause
        exit /b 1
    )
)

:: Verify the executable was created
if not exist "beanbot.exe" (
    echo ERROR: beanbot.exe was not created
    pause
    exit /b 1
)

echo.
echo ================================================
echo Setup complete! Starting BeanBot...
echo ================================================
echo.
echo USAGE TIPS:
echo - Describe your iTest issue in the text area
echo - BeanBot uses AI to provide troubleshooting steps
echo - Upload PDF documentation for enhanced context
echo - The knowledge base includes 216+ iTest documentation files
echo - Supports error codes, screenshots, and DrawIO diagrams
echo.
echo MODEL: gemma3:1b (optimized for technical troubleshooting)
echo.

:: Start the application
echo Starting BeanBot...
start "" "beanbot.exe"

:: Keep the window open briefly to show any startup messages
timeout /t 3 /nobreak >nul

echo BeanBot should now be running in a separate window.
echo If you encounter issues, check the console output above.
echo.
pause
