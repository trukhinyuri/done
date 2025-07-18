// Done - A gamified task manager application
// This is the main server application that provides REST API endpoints
// and serves the web interface for task management.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"done/lib/database"
	"done/lib/database/bolt"
	"done/lib/webview"
)

// Build-time variables injected by the build script
var (
	BuildVersion string = ""
	BuildTime    string = ""
)

// Command-line flags
var (
	dbUpgradePtr   *bool   // Flag to upgrade database schema
	servicePortPtr *int    // Port number for the HTTP server
	versionPtr     *bool   // Flag to display version information
	dbPathPtr      *string // Path to the database file
	nativePtr      *bool   // Flag to open in native window (macOS)
	chromePtr      *bool   // Flag to open in Chrome app mode
)

func init() {
	// Get home directory for default database path
	homeDir, err := os.UserHomeDir()
	defaultDBPath := "./tasks.db"
	if err == nil {
		defaultDBPath = filepath.Join(homeDir, "tasks.db")
	}
	
	dbUpgradePtr = flag.Bool("dbupgrade", false, "Upgrade database for new version compatibility")
	servicePortPtr = flag.Int("port", 3001, "Service port")
	versionPtr = flag.Bool("version", false, "Show app version")
	dbPathPtr = flag.String("dbpath", defaultDBPath, "Path to database file")
	nativePtr = flag.Bool("native", false, "Open in native window (macOS Safari app mode)")
	chromePtr = flag.Bool("chrome", false, "Open in Chrome app mode (macOS)")
}

func main() {
	flag.Parse()
	
	// Check if another instance is already running
	if !*versionPtr && isAlreadyRunning(*servicePortPtr) {
		log.Printf("Done is already running on port %d", *servicePortPtr)
		
		// If running as app, just open the browser/window
		if webview.IsRunningAsApp() {
			if *chromePtr {
				webview.LaunchWebViewChrome(*servicePortPtr)
			} else {
				webview.LaunchWebView(*servicePortPtr)
			}
			return
		}
		
		log.Fatal("Another instance is already running. Please close it first.")
	}
	
	submain(flag.Args())
}

// submain is the main entry point after flag parsing
// isAlreadyRunning checks if the application is already running on the given port
func isAlreadyRunning(port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func submain(args []string) {
	if *dbUpgradePtr == true {
		launchDBUpgrade()
	} else if *versionPtr == true {
		printServiceVersion()
	} else {
		startService()
	}
}

// launchDBUpgrade upgrades the database schema to the latest version
func launchDBUpgrade() {
	db := bolt.NewBoltDB(*dbPathPtr)

	err := db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Disconnect()

	result := db.DBUpgrade()
	log.Println(result)
}

// printServiceVersion displays the application version and build time
func printServiceVersion() {
	fmt.Println(BuildVersion)
	fmt.Println(BuildTime)
}

// startService initializes and starts the HTTP server
func startService() {
	db := bolt.NewBoltDB(*dbPathPtr)
	log.Println("Using local BoltDB:", *dbPathPtr)

	err := db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Disconnect()

	handler := database.NewHandler(db)

	// Set up HTTP routes
	mux := http.NewServeMux()

	// API endpoints
	apiPath := "/api"
	mux.HandleFunc(apiPath+"/version", testApi)                                                      // Get version information
	mux.HandleFunc(apiPath+"/getTasks", handler.GetTasks)                                           // Get all tasks
	mux.HandleFunc(apiPath+"/getTodayResults", handler.GetTodayResults)                             // Get today's completed tasks
	mux.HandleFunc(apiPath+"/addTask", handler.AddTask)                                             // Create a new task
	mux.HandleFunc(apiPath+"/removeTask", handler.RemoveTask)                                       // Delete a task
	mux.HandleFunc(apiPath+"/rearrangeTasks", handler.RearrangeTasks)                               // Reorder tasks (drag & drop)
	mux.HandleFunc(apiPath+"/completeTask", handler.CompleteTask)                                   // Mark task as completed
	mux.HandleFunc(apiPath+"/updateTaskExecutionRealSeconds", handler.UpdateTaskExecutionRealSeconds) // Update task timer
	mux.HandleFunc(apiPath+"/getGamification", handler.GetGamification)                             // Get gamification stats
	mux.HandleFunc(apiPath+"/updateGamification", handler.UpdateGamification)                       // Update gamification stats

	// Serve static files from frontend directory
	fileServer := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", http.StripPrefix("/", fileServer))

	servicePortString := strconv.Itoa(*servicePortPtr)
	log.Println("Starting server on :" + servicePortString)
	
	// Launch webview if requested or if running as .app bundle
	// But only if not launched by native launcher (which handles the UI)
	if (*nativePtr || *chromePtr) && !webview.IsRunningAsApp() {
		go func() {
			// Give the server a moment to start
			time.Sleep(500 * time.Millisecond)
			
			if *chromePtr {
				webview.LaunchWebViewChrome(*servicePortPtr)
			} else {
				webview.LaunchWebView(*servicePortPtr)
			}
		}()
	} else if !webview.IsRunningAsApp() {
		log.Printf("Open your browser and navigate to: http://localhost:%s", servicePortString)
	} else {
		log.Printf("Server started on port %s, waiting for native app to connect", servicePortString)
	}
	
	log.Fatal(http.ListenAndServe(":"+servicePortString, mux))
}

// Test struct is deprecated but kept for backward compatibility
type Test struct {
	Message  string
	CreateAt time.Time
}

// testApi returns version and build information for the application
func testApi(w http.ResponseWriter, r *http.Request) {
	// Generate build date if not set
	buildDate := BuildTime
	if buildDate == "" {
		buildDate = time.Now().Format("2006.01.02")
	} else {
		// Extract just the date part from BuildTime
		if t, err := time.Parse("2006-01-02 15:04:05", BuildTime); err == nil {
			buildDate = t.Format("2006.01.02")
		}
	}
	
	response := map[string]interface{}{
		"version": BuildVersion,
		"buildTime": BuildTime,
		"buildDate": buildDate,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
