# Mattermost Secrets Plugin

A Mattermost plugin for sending secret messages that can only be viewed once. This plugin provides a secure way to share sensitive information like passwords, API keys, or other private data that shouldn't persist in chat history.

## Features

- Send secret messages through Mattermost
- Messages can only be viewed once by each recipient
- Messages expire after a configurable time period
- Secure storage of sensitive information
- Supports slash commands for easy secret sharing
- Thread support for secret messages
- Multi-line secret support with formatting preservation
- Copy to clipboard functionality
- Automatic cleanup of expired secrets

## What is a Secret Message?

A secret message is a piece of text that:

- Can only be viewed once by each recipient
- Disappears after being viewed
- Has an expiration time (configurable)
- Is stored securely
- Can be part of a thread
- Supports formatted text (code blocks, multi-line content)

This is ideal for sending sensitive information like passwords, API keys, or other private data that shouldn't persist in chat history.

## Usage

### Sending a Secret Message

The easiest way to send a secret message is with the `/secret` slash command:

```
/secret Your temporary password is: P@ssw0rd123!
```

Anyone in the channel where you send the secret will see that a secret message exists, but only those who click the "View Secret" button will be able to see the actual content of the message.

#### Multi-line Secrets

The plugin supports multi-line secrets for sharing formatted content like code snippets, configuration files, or structured data. To create a multi-line secret:

1. Type `/secret` on the first line
2. Press Shift+Enter to start a new line
3. Type the content of your secret on subsequent lines

Example:
```
/secret
server: api.example.com
username: admin
password: s3cr3t
api_key: abcd1234efgh5678
```

This formatting is preserved when the secret is viewed and copied.

### Viewing a Secret Message

When someone sends a secret message, you'll see a post with a "View Secret" button. After clicking the button:

1. The secret message content will be displayed in a secure container
2. You can copy the content to your clipboard by clicking the copy icon
3. The secret will be marked as viewed for your user
4. Once you navigate away from the conversation, the secret will no longer be visible to you
5. If the secret is part of a thread, the viewing interface will appear in the thread context

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

After building, you'll have a file `dist/secrets-plugin-x.x.x.tar.gz`. Upload this plugin to your Mattermost instance:

1. Go to **System Console > Plugins > Plugin Management**
2. Upload the plugin file
3. Click "Enable" to activate the plugin

## Configuration

1. Go to **System Console > Plugins > Secrets Plugin**
2. Configure the following settings:
   - **Secret Expiry Time (minutes)**: Number of minutes before an unviewed secret expires (default: 60)
   - **Allow Copy to Clipboard**: Whether users can copy the secret to clipboard when viewing (default: true)

## Development

Detailed information for developers who want to understand, customize, or contribute to the plugin can be found in the [Development Guide](docs/development.md).

### Server

The server-side component is written in Go using the Mattermost Server plugin API. Key components:

- RESTful API endpoints for secret management
- Secure storage interface
- Slash command handling
- Automatic cleanup of expired secrets
- Post type customization

```bash
cd server
go test ./...
```

### Webapp

The webapp component is written in React using the Mattermost Webapp plugin API. Key features:

- Custom post type rendering
- Secret viewing interface
- Copy to clipboard functionality
- Integration with Mattermost's UI

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
- Sharing formatted code snippets or configuration files
- Thread-based secret discussions

## Security Considerations

- Secrets are stored securely and can only be viewed once
- Expired secrets are automatically cleaned up
- Secret content is only transmitted to authorized users
- The plugin respects Mattermost's permission system
- Secret viewing is tracked per user

## License

Apache License 2.0