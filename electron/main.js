const { app, BrowserWindow, shell } = require('electron');
const { spawn } = require('child_process');
const path = require('path');
const net = require('net');

let mainWindow;
let serverProcess;

// Function to check if port is available
function checkPort(port) {
  return new Promise((resolve) => {
    const server = net.createServer();
    server.once('error', () => resolve(false));
    server.once('listening', () => {
      server.close();
      resolve(true);
    });
    server.listen(port, 'localhost');
  });
}

// Function to find an available port
async function findAvailablePort(startPort = 3001) {
  let port = startPort;
  while (!(await checkPort(port))) {
    port++;
  }
  return port;
}

// Function to wait for server to be ready
function waitForServer(port, maxAttempts = 30) {
  return new Promise((resolve, reject) => {
    let attempts = 0;
    
    const tryConnect = () => {
      const client = new net.Socket();
      
      client.connect(port, 'localhost', () => {
        client.end();
        resolve();
      });
      
      client.on('error', () => {
        attempts++;
        if (attempts >= maxAttempts) {
          reject(new Error('Server failed to start'));
        } else {
          setTimeout(tryConnect, 100);
        }
      });
    };
    
    tryConnect();
  });
}

function createWindow(port) {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true
    },
    title: 'Done - Task Manager',
    icon: path.join(__dirname, 'icon.png') // You'll need to add an icon
  });

  mainWindow.loadURL(`http://localhost:${port}`);

  // Open external links in default browser and prevent external navigation
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    if (url.startsWith(`http://localhost:${port}`)) {
      return { action: 'allow' };
    }
    shell.openExternal(url);
    return { action: 'deny' };
  });

  mainWindow.webContents.on('will-navigate', (event, url) => {
    if (!url.startsWith(`http://localhost:${port}`)) {
      event.preventDefault();
      shell.openExternal(url);
    }
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

app.whenReady().then(async () => {
  // Find an available port
  const port = await findAvailablePort();
  
  // Start the Done server
  const doneExecutable = app.isPackaged 
    ? path.join(process.resourcesPath, 'done')
    : path.join(__dirname, '..', 'done');
    
  const frontendPath = app.isPackaged
    ? process.resourcesPath
    : path.join(__dirname, '..');
  
  serverProcess = spawn(doneExecutable, ['-port', port.toString()], {
    cwd: frontendPath,
    stdio: 'inherit'
  });

  serverProcess.on('error', (err) => {
    console.error('Failed to start server:', err);
    app.quit();
  });

  // Wait for server to be ready
  try {
    await waitForServer(port);
    createWindow(port);
  } catch (error) {
    console.error('Server failed to start:', error);
    app.quit();
  }
});

app.on('window-all-closed', () => {
  if (serverProcess) {
    serverProcess.kill();
  }
  app.quit();
});

app.on('activate', () => {
  if (mainWindow === null) {
    createWindow();
  }
});

// Clean up server process on app quit
app.on('before-quit', () => {
  if (serverProcess) {
    serverProcess.kill();
  }
});
