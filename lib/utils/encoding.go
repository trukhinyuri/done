package utils

import (
	"encoding/base64"
	"strings"
)

// DecodeTaskText decodes task text that may be encoded in various formats
func DecodeTaskText(text string) string {
	if text == "" {
		return text
	}

	// Try base64 decoding first (URL-safe variant)
	if isBase64Encoded(text) {
		decoded, err := decodeBase64(text)
		if err == nil {
			return decoded
		}
	}

	// Handle legacy UUID format for backward compatibility
	if strings.Contains(text, "280d382c-f23e-4631-8551-f43661405497") ||
		strings.Contains(text, "e6f23f57-6cad-451b-8306-7939e25542dc") ||
		strings.Contains(text, "a7f3d0a1-2b5e-4c6d-8e9f-1a2b3c4d5e6f") {
		
		text = strings.ReplaceAll(text, "280d382c-f23e-4631-8551-f43661405497", "\"")
		text = strings.ReplaceAll(text, "e6f23f57-6cad-451b-8306-7939e25542dc", "'")
		text = strings.ReplaceAll(text, "a7f3d0a1-2b5e-4c6d-8e9f-1a2b3c4d5e6f", "\n")
		return text
	}

	// Handle simple escaped format
	if strings.Contains(text, "\\") {
		text = strings.ReplaceAll(text, "\\n", "\n")
		text = strings.ReplaceAll(text, "\\r", "\r")
		text = strings.ReplaceAll(text, "\\t", "\t")
		text = strings.ReplaceAll(text, "\\'", "'")
		text = strings.ReplaceAll(text, "\\\"", "\"")
		text = strings.ReplaceAll(text, "\\\\", "\\")
		return text
	}

	// Return as-is if no encoding detected
	return text
}

// CleanTaskText removes any encoding artifacts from task text
func CleanTaskText(text string) string {
	// First decode
	text = DecodeTaskText(text)
	
	// Remove any remaining UUID artifacts (in case of double encoding)
	text = strings.ReplaceAll(text, "280d382c-f23e-4631-8551-f43661405497", "\"")
	text = strings.ReplaceAll(text, "e6f23f57-6cad-451b-8306-7939e25542dc", "'")
	text = strings.ReplaceAll(text, "a7f3d0a1-2b5e-4c6d-8e9f-1a2b3c4d5e6f", "\n")
	
	return text
}

func isBase64Encoded(s string) bool {
	// Check if string only contains base64 characters (including URL-safe variants)
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || 
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return len(s) > 0
}

func decodeBase64(s string) (string, error) {
	// Convert URL-safe base64 to standard base64
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	
	return string(decoded), nil
}