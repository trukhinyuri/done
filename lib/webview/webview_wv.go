// +build wv

// This file provides a native webview implementation using github.com/webview/webview
// To enable this implementation, build with: go build -tags wv
// First install the dependency: go get github.com/webview/webview

package webview

import (
	"fmt"
	"log"
	"net"
	"time"
	
	"github.com/webview/webview"
)

// LaunchWebView creates a native webview window
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
	
	// Create webview
	w := webview.New(true)
	defer w.Destroy()
	
	w.SetTitle("Done - Task Manager")
	w.SetSize(1200, 800, webview.HintNone)
	w.Navigate(url)
	
	// Run the webview (this blocks until window is closed)
	w.Run()
}

// LaunchWebViewChrome is the same as LaunchWebView when using native webview
func LaunchWebViewChrome(port int) {
	LaunchWebView(port)
}

// IsRunningAsApp always returns false when using webview library
func IsRunningAsApp() bool {
	return false
}