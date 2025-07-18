# Done Electron App

This directory contains an Electron wrapper for the Done task manager, providing a true native application experience.

## Prerequisites

- Node.js and npm installed
- The Done binary built in the parent directory

## Setup

1. First, build the Done binary:
   ```bash
   cd ..
   go build -o done
   ```

2. Install Electron dependencies:
   ```bash
   cd electron
   npm install
   ```

## Running in Development

```bash
npm start
```

## Building the App

To build a distributable macOS app:

```bash
npm run dist
```

The built app will be in the `dist` folder.

## Features

- Automatic port selection (finds an available port if 3001 is in use)
- Server process management (starts and stops with the app)
- Native window with standard macOS controls
- No browser UI elements

## Notes

- The Electron app embeds and manages the Done server process
- When you close the Electron window, the server automatically stops
- All data is stored in the same location as when running Done directly