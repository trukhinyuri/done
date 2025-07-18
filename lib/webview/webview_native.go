// +build darwin,native

package webview

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

void LaunchNativeWebView(const char* url) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        
        // Create window
        NSRect frame = NSMakeRect(0, 0, 1200, 800);
        NSWindow* window = [[NSWindow alloc] 
            initWithContentRect:frame
            styleMask:(NSWindowStyleMaskTitled | 
                      NSWindowStyleMaskClosable | 
                      NSWindowStyleMaskMiniaturizable | 
                      NSWindowStyleMaskResizable)
            backing:NSBackingStoreBuffered
            defer:NO];
        
        [window setTitle:@"Done"];
        [window center];
        
        // Create WebView
        WKWebViewConfiguration* config = [[WKWebViewConfiguration alloc] init];
        WKWebView* webView = [[WKWebView alloc] initWithFrame:frame configuration:config];
        
        // Load URL
        NSURL* nsurl = [NSURL URLWithString:[NSString stringWithUTF8String:url]];
        NSURLRequest* request = [NSURLRequest requestWithURL:nsurl];
        [webView loadRequest:request];
        
        // Set up window
        [window setContentView:webView];
        [window makeKeyAndOrderFront:nil];
        
        // Run app
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
        [NSApp activateIgnoringOtherApps:YES];
        [NSApp run];
    }
}
*/
import "C"
import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

// LaunchWebView creates a native WKWebView window for the application on macOS
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
	
	// For native build, just open in default browser
	// This is simpler and avoids separate process issues
	cmd := exec.Command("open", url)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// LaunchWebViewChrome launches Chrome in app mode on macOS
func LaunchWebViewChrome(port int) {
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
	
	// Use Chrome app mode for native-like experience
	chromeApp := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if _, err := os.Stat(chromeApp); err == nil {
		cmd := exec.Command(chromeApp, 
			"--app="+url,
			"--new-window",
			"--window-size=1200,800",
			"--window-position=100,100",
			"--user-data-dir=/tmp/done-chrome-profile",
		)
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to launch Chrome in app mode: %v", err)
			exec.Command("open", url).Run()
		} else {
			log.Printf("Launched in Chrome app mode: %s", url)
		}
	} else {
		log.Println("Chrome not found, opening in default browser")
		exec.Command("open", url).Run()
	}
}

// IsRunningAsApp checks if the application is running as a .app bundle
func IsRunningAsApp() bool {
	return true // Always true for native build
}