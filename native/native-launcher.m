#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

@interface AppDelegate : NSObject <NSApplicationDelegate, WKNavigationDelegate>
@property (nonatomic, strong) NSWindow *window;
@property (nonatomic, strong) WKWebView *webView;
@property (nonatomic, strong) NSTask *serverTask;
@end

@implementation AppDelegate

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    // Start the server
    [self startServer];
    
    // Wait a moment for server to start
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 1 * NSEC_PER_SEC), dispatch_get_main_queue(), ^{
        [self createWindow];
    });
}

- (void)startServer {
    NSString *appPath = [[NSBundle mainBundle] bundlePath];
    NSString *serverPath = [appPath stringByAppendingPathComponent:@"Contents/MacOS/done"];
    NSString *resourcesPath = [appPath stringByAppendingPathComponent:@"Contents/Resources"];
    
    self.serverTask = [[NSTask alloc] init];
    self.serverTask.launchPath = serverPath;
    self.serverTask.currentDirectoryPath = resourcesPath;
    
    NSLog(@"Starting server: %@", serverPath);
    NSLog(@"Working directory: %@", resourcesPath);
    
    [self.serverTask launch];
}

- (void)createWindow {
    // Create window
    NSRect windowFrame = NSMakeRect(100, 100, 1200, 800);
    NSWindowStyleMask styleMask = NSWindowStyleMaskTitled | 
                                  NSWindowStyleMaskClosable | 
                                  NSWindowStyleMaskMiniaturizable | 
                                  NSWindowStyleMaskResizable;
    
    self.window = [[NSWindow alloc] initWithContentRect:windowFrame
                                              styleMask:styleMask
                                                backing:NSBackingStoreBuffered
                                                  defer:NO];
    
    self.window.title = @"Done";
    self.window.minSize = NSMakeSize(800, 600);
    
    // Create WebView
    WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];
    config.preferences.javaScriptEnabled = YES;
    
    self.webView = [[WKWebView alloc] initWithFrame:windowFrame configuration:config];
    self.webView.navigationDelegate = self;
    
    // Load the app
    NSURL *url = [NSURL URLWithString:@"http://localhost:3001"];
    NSURLRequest *request = [NSURLRequest requestWithURL:url];
    [self.webView loadRequest:request];
    
    // Set up window
    self.window.contentView = self.webView;
    [self.window center];
    [self.window makeKeyAndOrderFront:nil];
    
    // Keep app running
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
    [NSApp activateIgnoringOtherApps:YES];
}

- (void)webView:(WKWebView *)webView didFailNavigation:(WKNavigation *)navigation withError:(NSError *)error {
    NSLog(@"Failed to load: %@", error.localizedDescription);
    
    // Retry after a delay
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 2 * NSEC_PER_SEC), dispatch_get_main_queue(), ^{
        NSURL *url = [NSURL URLWithString:@"http://localhost:3001"];
        NSURLRequest *request = [NSURLRequest requestWithURL:url];
        [self.webView loadRequest:request];
    });
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
    return YES;
}

- (void)applicationWillTerminate:(NSNotification *)notification {
    if (self.serverTask && self.serverTask.isRunning) {
        [self.serverTask terminate];
        [self.serverTask waitUntilExit];
    }
}

@end

int main(int argc, const char * argv[]) {
    @autoreleasepool {
        NSApplication *app = [NSApplication sharedApplication];
        AppDelegate *delegate = [[AppDelegate alloc] init];
        app.delegate = delegate;
        
        return NSApplicationMain(argc, argv);
    }
}