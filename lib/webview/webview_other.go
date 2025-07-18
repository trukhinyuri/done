// +build !darwin

package webview

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// LaunchWebView opens the application in the default browser on non-macOS platforms
func LaunchWebView(port int) {
	url := fmt.Sprintf("http://localhost:%d", port)
	
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		log.Printf("Please open your browser and navigate to: %s", url)
		return
	}
	
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to open browser: %v", err)
		log.Printf("Please open your browser and navigate to: %s", url)
	}
}

// LaunchWebViewChrome is not implemented on non-macOS platforms
func LaunchWebViewChrome(port int) {
	LaunchWebView(port)
}

// IsRunningAsApp always returns false on non-macOS platforms
func IsRunningAsApp() bool {
	return false
}