# Mattermost Secrets Plugin

A Mattermost plugin for sending secret messages that can only be viewed once.

## Features

- Send secret messages through Mattermost
- Messages can only be viewed once
- Messages expire after a configurable time period
- Supports both slash commands and UI options

## Screenshots

![Secret Message](assets/screenshot1.png)

## Building the Plugin

### Prerequisites

- Go 1.20 or later
- Node.js 16.x or later
- npm 7.x or later

### Build Process

#### Option 1: Using Makefile (recommended)

```bash
# Clean previous builds
make clean

# Build the plugin
make build

# Create the distribution bundle
make dist
```

#### Option 2: Using Node.js build script

```bash
# Install Node.js dependencies and run the build script
node build.js
```

#### Option 3: Manual build

1. Build the server component:
   ```bash
   cd server
   go mod tidy
   go build -o ../build/server/dist/plugin-windows-amd64.exe
   ```

2. Build the webapp component:
   ```bash
   cd webapp
   npm install --legacy-peer-deps
   npm run build
   ```

3. Copy files to the build directory:
   ```bash
   cp plugin.json build/
   mkdir -p build/webapp
   cp -r webapp/dist build/webapp/
   ```

## Installing the Plugin

After building, you'll have a file `dist/com.mattermost.secrets-plugin-0.1.0.tar.gz`. Upload this plugin to your Mattermost instance:

1. Go to **System Console > Plugins > Plugin Management**
2. Upload the plugin file
3. Click "Enable" to activate the plugin

## Configuration

1. Go to **System Console > Plugins > Secrets Plugin**
2. Configure the following settings:
   - **Secret Expiry Time (hours)**: Number of hours before an unviewed secret expires
   - **Allow Copy to Clipboard**: Whether users can copy the secret to clipboard

## Usage

### Slash Command

Send a secret message using the slash command:

```
/secret This is a secret message
```

### UI Button

1. Click the "+" button in the message input
2. Select "Send Secret Message"
3. Enter your secret message
4. Click "Send"

## Development

### Server

The server-side component is written in Go using the Mattermost Server plugin API.

```bash
cd server
go test ./...
```

### Webapp

The webapp component is written in React using the Mattermost Webapp plugin API.

```bash
cd webapp
npm run test
npm run lint
```

## License

Apache License 2.0