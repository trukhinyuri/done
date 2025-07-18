// +build darwin,!native

package webview

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// LaunchWebView creates a native window for the application on macOS
// This version prefers Safari for better integration with the app
func LaunchWebView(port int) {
	url := fmt.Sprintf("http://localhost:%d", port)
	
	// Wait for server to be ready
	log.Println("Waiting for server to start...")
	for i := 0; i < 30; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	// When running as .app, just open in default browser
	// This ensures the browser window is associated with our app
	if IsRunningAsApp() {
		// Simply open URL in default browser
		cmd := exec.Command("open", url)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to open browser: %v", err)
		}
		return
	}
	
	// For command line usage, try Safari with hidden toolbar
	script := fmt.Sprintf(`
tell application "Safari"
	activate
	open location "%s"
	delay 0.5
	tell application "System Events"
		tell process "Safari"
			try
				-- Hide toolbar
				keystroke "t" using {command down, option down}
			end try
		end tell
	end tell
end tell
`, url)
	
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		// Fallback: open in default browser
		log.Printf("Failed to open Safari: %v", err)
		exec.Command("open", url).Run()
	}
}

// LaunchWebViewChrome launches the app in a Chrome app window
func LaunchWebViewChrome(port int) {
	// For .app bundle, just use regular LaunchWebView
	// This avoids creating separate Chrome processes
	LaunchWebView(port)
}

// IsRunningAsApp checks if the application is running as a .app bundle
func IsRunningAsApp() bool {
	// Check if we're running from within a .app bundle
	executable, _ := os.Executable()
	return strings.Contains(executable, ".app/Contents/MacOS")
}