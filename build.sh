#!/usr/bin/env bash

# Exit on error, undefined variable, and pipe failures
set -euo pipefail

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Default values
VERSION="${VERSION:-0.1}"
BUILD_TIME=$(date +"%Y-%m-%d %T")
OUTPUT_NAME="done"
LDFLAGS=""
VERBOSE=false
CLEAN=false
ALL_PLATFORMS=false
BUILD_APP=false

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[+]${NC} $1"
}

print_error() {
    echo -e "${RED}[!]${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}[*]${NC} $1"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build script for the Done application

OPTIONS:
    -h, --help          Show this help message
    -v, --version       Set version (default: $VERSION)
    -o, --output        Set output binary name (default: $OUTPUT_NAME)
    -p, --platform      Target platform (e.g., linux/amd64, darwin/amd64, windows/amd64)
    -a, --all           Build for all major platforms
    -c, --clean         Clean build artifacts before building
    -V, --verbose       Enable verbose output
    -s, --static        Build static binary (CGO_ENABLED=0)
    -d, --docker        Build using Docker
    --ldflags           Additional ldflags
    --app               Build macOS .app bundle (macOS only)

EXAMPLES:
    $0                              # Build for current platform (includes .app on macOS)
    $0 -v 1.0.0                     # Build with version 1.0.0
    $0 -p linux/amd64               # Build for Linux AMD64
    $0 -a                           # Build for all platforms
    $0 -c -s                        # Clean and build static binary
    $0 -d                           # Build using Docker
    $0 --app                        # Build only macOS .app bundle

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_NAME="$2"
            shift 2
            ;;
        -p|--platform)
            TARGET_PLATFORM="$2"
            shift 2
            ;;
        -a|--all)
            ALL_PLATFORMS=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -V|--verbose)
            VERBOSE=true
            shift
            ;;
        -s|--static)
            export CGO_ENABLED=0
            shift
            ;;
        -d|--docker)
            DOCKER_BUILD=true
            shift
            ;;
        --ldflags)
            LDFLAGS="$2"
            shift 2
            ;;
        --app)
            BUILD_APP=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Enable verbose mode if requested
if [ "$VERBOSE" = true ]; then
    set -x
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

# Show Go version
GO_VERSION=$(go version)
print_status "Using $GO_VERSION"

# Clean if requested
if [ "$CLEAN" = true ]; then
    print_status "Cleaning build artifacts..."
    rm -rf bin/
    rm -f "$OUTPUT_NAME"
    rm -f "${OUTPUT_NAME}.exe"
    go clean -cache
fi

# Create bin directory if it doesn't exist
mkdir -p bin

# Build flags
BUILD_FLAGS="-ldflags=\"-X 'main.BuildTime=$BUILD_TIME' -X 'main.BuildVersion=$VERSION' $LDFLAGS\""

# Function to build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT=$3
    
    print_status "Building for $GOOS/$GOARCH..."
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "bin/${OUTPUT}" -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.BuildVersion=$VERSION' $LDFLAGS" .
    
    if [ $? -eq 0 ]; then
        print_status "Successfully built: bin/${OUTPUT}"
    else
        print_error "Failed to build for $GOOS/$GOARCH"
        exit 1
    fi
}

# Docker build function
docker_build() {
    print_status "Building with Docker..."
    docker build -t "${OUTPUT_NAME}:${VERSION}" -t "${OUTPUT_NAME}:latest" .
    if [ $? -eq 0 ]; then
        print_status "Docker build successful: ${OUTPUT_NAME}:${VERSION}"
    else
        print_error "Docker build failed"
        exit 1
    fi
}

# macOS .app bundle build function
build_macos_app() {
    if [ "$(uname)" != "Darwin" ]; then
        print_error "Building .app bundle is only supported on macOS"
        exit 1
    fi
    
    print_status "Building macOS .app bundle..."
    
    # Check if binary already exists (from previous build step)
    if [ ! -f "$OUTPUT_NAME" ]; then
        # First build the binary
        print_status "Building macOS binary..."
        env GOOS="darwin" GOARCH="$(uname -m | sed 's/x86_64/amd64/;s/arm64/arm64/')" go build -o "$OUTPUT_NAME" -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.BuildVersion=$VERSION' $LDFLAGS" .
        
        if [ $? -ne 0 ]; then
            print_error "Failed to build macOS binary"
            exit 1
        fi
    fi
    
    # Check if native version already exists, if not try to build it
    if [ ! -f "${OUTPUT_NAME}-native" ]; then
        # Try to build native version for better experience
        print_status "Attempting to build native WebView version..."
        env GOOS="darwin" GOARCH="$(uname -m | sed 's/x86_64/amd64/;s/arm64/arm64/')" go build -tags native -o "${OUTPUT_NAME}-native" -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.BuildVersion=$VERSION' $LDFLAGS" . 2>/dev/null
    fi
    
    # Use native version if available for app bundle
    BINARY_FOR_APP="$OUTPUT_NAME"
    if [ -f "${OUTPUT_NAME}-native" ]; then
        print_status "Using native WebView build for .app bundle"
        BINARY_FOR_APP="${OUTPUT_NAME}-native"
    else
        print_warning "Native build not available, using standard build for .app bundle"
    fi
    
    # Create app bundle structure
    APP_NAME="Done.app"
    APP_DIR="bin/$APP_NAME"
    
    print_status "Creating app bundle structure..."
    rm -rf "$APP_DIR"
    mkdir -p "$APP_DIR/Contents/MacOS"
    mkdir -p "$APP_DIR/Contents/Resources"
    
    # Copy binary
    cp "$BINARY_FOR_APP" "$APP_DIR/Contents/MacOS/done"
    chmod +x "$APP_DIR/Contents/MacOS/done"
    
    # Compile native launcher
    print_status "Compiling native launcher..."
    clang -o "$APP_DIR/Contents/MacOS/done-launcher" native/native-launcher.m \
        -framework Cocoa -framework WebKit \
        -mmacosx-version-min=10.13 \
        || {
            print_warning "Failed to compile native launcher, using shell script fallback"
            NATIVE_LAUNCHER_FAILED=true
        }
    
    # Copy frontend directory
    cp -r frontend "$APP_DIR/Contents/Resources/"
    
    # Copy icon if it exists
    if [ -f "assets/icon.icns" ]; then
        cp assets/icon.icns "$APP_DIR/Contents/Resources/"
        print_status "Icon copied to app bundle"
    elif [ -f "assets/icon.svg" ]; then
        # If only SVG exists, copy it as a placeholder
        cp assets/icon.svg "$APP_DIR/Contents/Resources/"
        print_status "SVG icon copied to app bundle"
    fi
    
    # Create Info.plist
    cat > "$APP_DIR/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleDisplayName</key>
    <string>Done</string>
    <key>CFBundleExecutable</key>
    <string>done-launcher</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>CFBundleIdentifier</key>
    <string>com.trukhinyuri.done</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>Done</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.productivity</string>
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsLocalNetworking</key>
        <true/>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>
    <key>NSMainNibFile</key>
    <string></string>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
</dict>
</plist>
EOF
    
    # Create launcher script only if native launcher failed
    if [ "${NATIVE_LAUNCHER_FAILED:-false}" = true ]; then
        print_status "Creating shell script launcher..."
        cat > "$APP_DIR/Contents/MacOS/done-launcher" << 'EOF'
#!/bin/bash
# Get the directory where the app bundle is located
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RESOURCES_DIR="$SCRIPT_DIR/../Resources"
LOG_FILE="$HOME/Library/Logs/Done.log"

# Create log directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"

# Log startup
echo "[$(date)] Starting Done.app from $SCRIPT_DIR" >> "$LOG_FILE"

# Change to Resources directory where frontend is located
cd "$RESOURCES_DIR" || {
    echo "[$(date)] ERROR: Failed to change to Resources directory: $RESOURCES_DIR" >> "$LOG_FILE"
    exit 1
}

# Log current directory
echo "[$(date)] Working directory: $(pwd)" >> "$LOG_FILE"

# Launch the actual binary with frontend in current directory
# The app will detect it's running as a .app bundle and open in native window
# Redirect output to log file and run in background to prevent launcher hanging
"$SCRIPT_DIR/done" "$@" >> "$LOG_FILE" 2>&1 &

# Give the process a moment to start
sleep 1

# Check if the process started successfully
if pgrep -f "$SCRIPT_DIR/done" > /dev/null; then
    echo "[$(date)] Done started successfully" >> "$LOG_FILE"
else
    echo "[$(date)] ERROR: Failed to start Done" >> "$LOG_FILE"
    exit 1
fi
EOF
        chmod +x "$APP_DIR/Contents/MacOS/done-launcher"
    fi
    
    # Create a simple icon (you can replace this with a proper .icns file)
    # For now, we'll just touch the file
    touch "$APP_DIR/Contents/Resources/icon.icns"
    
    print_status "macOS app bundle created: $APP_DIR"
    print_status "You can run it with: open $APP_DIR"
    
    # Don't clean up binaries when called from main build process
    # Only clean when explicitly building just the app bundle
    if [ "${BUILD_APP}" = true ]; then
        # Clean up the standalone binary only when explicitly building app
        rm -f "$OUTPUT_NAME" "${OUTPUT_NAME}-native"
    fi
}

# Main build logic
if [ "${DOCKER_BUILD:-false}" = true ]; then
    docker_build
elif [ "$BUILD_APP" = true ]; then
    build_macos_app
elif [ "$ALL_PLATFORMS" = true ]; then
    # Build for all major platforms
    build_platform "linux" "amd64" "${OUTPUT_NAME}-linux-amd64"
    build_platform "linux" "arm64" "${OUTPUT_NAME}-linux-arm64"
    build_platform "darwin" "amd64" "${OUTPUT_NAME}-darwin-amd64"
    build_platform "darwin" "arm64" "${OUTPUT_NAME}-darwin-arm64"
    build_platform "windows" "amd64" "${OUTPUT_NAME}-windows-amd64"
elif [ -n "${TARGET_PLATFORM:-}" ]; then
    # Build for specific platform
    IFS='/' read -r GOOS GOARCH <<< "$TARGET_PLATFORM"
    build_platform "$GOOS" "$GOARCH" "${OUTPUT_NAME}-${GOOS}-${GOARCH}"
else
    # Build for current platform
    CURRENT_OS=$(go env GOOS)
    CURRENT_ARCH=$(go env GOARCH)
    print_status "Building for current platform ($CURRENT_OS/$CURRENT_ARCH)..."
    
    go build -o "$OUTPUT_NAME" -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.BuildVersion=$VERSION' $LDFLAGS" .
    
    if [ $? -eq 0 ]; then
        print_status "Build successful: $OUTPUT_NAME"
        print_status "Version: $VERSION"
        print_status "Build time: $BUILD_TIME"
        
        # Show binary info
        if command -v file &> /dev/null; then
            file "$OUTPUT_NAME"
        fi
        
        # On macOS, also build native WebView version and .app bundle
        if [ "$CURRENT_OS" = "darwin" ]; then
            print_status "Building native WebView version for macOS..."
            go build -tags native -o "${OUTPUT_NAME}-native" -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.BuildVersion=$VERSION' $LDFLAGS" .
            
            if [ $? -eq 0 ]; then
                print_status "Native build successful: ${OUTPUT_NAME}-native"
                if command -v file &> /dev/null; then
                    file "${OUTPUT_NAME}-native"
                fi
            else
                print_warning "Native build failed, continuing with standard build"
            fi
            
            # Also build .app bundle automatically on macOS
            print_status "Building .app bundle automatically..."
            build_macos_app
        fi
    else
        print_error "Build failed"
        exit 1
    fi
fi

# Run tests if not cross-compiling
if [ -z "${TARGET_PLATFORM:-}" ] && [ "$ALL_PLATFORMS" != true ]; then
    print_status "Running tests..."
    go test ./... -v
    if [ $? -eq 0 ]; then
        print_status "All tests passed"
    else
        print_warning "Some tests failed"
    fi
fi

print_status "Build complete!"