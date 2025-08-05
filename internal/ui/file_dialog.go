package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Windows API constants for file dialog
const (
	OFN_ALLOWMULTISELECT = 0x00000200
	OFN_EXPLORER         = 0x00080000
	OFN_FILEMUSTEXIST    = 0x00001000
	OFN_HIDEREADONLY     = 0x00000004
	OFN_LONGNAMES        = 0x00200000
	OFN_NOCHANGEDIR      = 0x00000008
	OFN_PATHMUSTEXIST    = 0x00000800
)

// OPENFILENAME structure for Windows file dialog
type OPENFILENAME struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
}

var (
	comdlg32            = syscall.NewLazyDLL("comdlg32.dll")
	getOpenFileNameProc = comdlg32.NewProc("GetOpenFileNameW")
)

// ShowFileDialog opens the Windows system file dialog and returns selected file paths
func ShowFileDialog() ([]string, error) {
	// Create filter string for supported file types using UTF-16 directly
	filterParts := []string{
		"All Supported Files",
		"*.txt;*.log;*.logs;*.pdf;*.docx;*.html;*.png;*.jpg;*.jpeg;*.bmp;*.gif;*.tiff",
		"Text Files (*.txt)",
		"*.txt",
		"Log Files (*.log, *.logs)",
		"*.log;*.logs",
		"PDF Files (*.pdf)",
		"*.pdf",
		"Word Documents (*.docx)",
		"*.docx",
		"HTML Files (*.html)",
		"*.html",
		"Image Files",
		"*.png;*.jpg;*.jpeg;*.bmp;*.gif;*.tiff",
		"All Files (*.*)",
		"*.*",
		"", // Final empty string to terminate
	}

	// Convert to UTF-16 and build filter buffer
	var filterBuffer []uint16
	for _, part := range filterParts {
		utf16Part, err := syscall.UTF16FromString(part)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter part '%s' to UTF-16: %w", part, err)
		}
		filterBuffer = append(filterBuffer, utf16Part...)
	}
	// Add final null terminator
	filterBuffer = append(filterBuffer, 0)

	// Create buffer for file path (support multiple selection)
	const maxPath = 32768
	fileBuffer := make([]uint16, maxPath)

	// Create title for dialog
	title := "Select files to upload"
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return nil, fmt.Errorf("failed to convert title to UTF-16: %w", err)
	}

	// Get current working directory for initial directory
	cwd, _ := os.Getwd()
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)

	// Setup OPENFILENAME structure
	ofn := OPENFILENAME{
		lStructSize:     uint32(unsafe.Sizeof(OPENFILENAME{})),
		hwndOwner:       0,
		hInstance:       0,
		lpstrFilter:     &filterBuffer[0],
		nFilterIndex:    1,
		lpstrFile:       &fileBuffer[0],
		nMaxFile:        maxPath,
		lpstrInitialDir: cwdPtr,
		lpstrTitle:      titlePtr,
		flags: OFN_ALLOWMULTISELECT | OFN_EXPLORER | OFN_FILEMUSTEXIST |
			OFN_HIDEREADONLY | OFN_LONGNAMES | OFN_NOCHANGEDIR | OFN_PATHMUSTEXIST,
	}

	// Call the Windows API
	ret, _, _ := getOpenFileNameProc.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		// User cancelled or error occurred
		return nil, nil
	}

	// Parse the result
	files := parseFileBuffer(fileBuffer)

	// Validate that files exist
	var validFiles []string
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			validFiles = append(validFiles, file)
		}
	}

	return validFiles, nil
}

// parseFileBuffer parses the file buffer returned by GetOpenFileName
func parseFileBuffer(buffer []uint16) []string {
	var files []string

	// Convert buffer to string
	str := syscall.UTF16ToString(buffer)
	if str == "" {
		return files
	}

	// Find the directory and file names
	// Format: "directory\0file1\0file2\0\0" for multiple files
	// Format: "fullpath\0\0" for single file

	parts := splitNullTerminated(buffer)
	if len(parts) == 0 {
		return files
	}

	if len(parts) == 1 {
		// Single file selection
		files = append(files, parts[0])
	} else {
		// Multiple file selection
		directory := parts[0]
		for i := 1; i < len(parts); i++ {
			if parts[i] != "" {
				fullPath := filepath.Join(directory, parts[i])
				files = append(files, fullPath)
			}
		}
	}

	return files
}

// splitNullTerminated splits a UTF-16 buffer by null terminators
func splitNullTerminated(buffer []uint16) []string {
	var parts []string
	var current []uint16

	for i := 0; i < len(buffer); i++ {
		if buffer[i] == 0 {
			if len(current) > 0 {
				parts = append(parts, syscall.UTF16ToString(current))
				current = nil
			} else if len(parts) > 0 {
				// Double null terminator found, end of data
				break
			}
		} else {
			current = append(current, buffer[i])
		}
	}

	// Add the last part if it doesn't end with null
	if len(current) > 0 {
		parts = append(parts, syscall.UTF16ToString(current))
	}

	return parts
}
