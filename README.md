# Mattermost Secrets Plugin

A Mattermost plugin for sending secret messages that can only be viewed once.

## Features

- Send secret messages through Mattermost
- Messages can only be viewed once by each recipient
- Messages expire after a configurable time period
- Secure storage of sensitive information
- Supports slash commands for easy secret sharing

## What is a Secret Message?

A secret message is a piece of text that:

- Can only be viewed once by each recipient
- Disappears after being viewed
- Has an expiration time
- Is stored securely

This is ideal for sending sensitive information like passwords, API keys, or other private data that shouldn't persist in chat history.

## Usage

### Sending a Secret Message

The easiest way to send a secret message is with the `/secret` slash command:

```
/secret Your temporary password is: P@ssw0rd123!
```

Anyone in the channel where you send the secret will see that a secret message exists, but only those who click the "View Secret" button will be able to see the actual content of the message.

### Viewing a Secret Message

When someone sends a secret message, you'll see a post with a "View Secret" button. After clicking the button, the secret message content will be displayed in a secure container. You can copy the content to your clipboard by clicking the copy icon.

Once you navigate away from the conversation, the secret will no longer be visible to you.

For more detailed usage instructions, see the [User Guide](docs/user_guide.md).

## Building the Plugin

### Prerequisites

- Go 1.20 or later
- Node.js 16.x or later
- npm 7.x or later

### Build Process

```bash
# Clean previous builds
make clean

# Build the plugin
make build

# Create the distribution bundle
make dist
```

## Installing the Plugin

After building, you'll have a file `dist/com.mattermost.secrets-plugin-0.1.0.tar.gz`. Upload this plugin to your Mattermost instance:

1. Go to **System Console > Plugins > Plugin Management**
2. Upload the plugin file
3. Click "Enable" to activate the plugin

## Configuration

1. Go to **System Console > Plugins > Secrets Plugin**
2. Configure the following settings:
   - **Secret Expiry Time (hours)**: Number of hours before an unviewed secret expires (default: 24)
   - **Allow Copy to Clipboard**: Whether users can copy the secret to clipboard when viewing (default: true)

## Development

Detailed information for developers who want to understand, customize, or contribute to the plugin can be found in the [Development Guide](docs/development.md).

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

## Use Cases

The Secrets Plugin is ideal for:

- Sharing temporary passwords
- Sending API keys or tokens
- Communicating sensitive personal information
- Sharing confidential URLs or access information

## License

Apache License 2.0