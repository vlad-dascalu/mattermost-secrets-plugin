# Mattermost Secrets Plugin

A Mattermost plugin for sending secret messages that can only be viewed once by each recipient.

## Features

- Send messages that are only viewable once by each recipient
- Automatic expiration of unviewed secrets
- Copy message contents to clipboard
- Secure storage of secret messages in plugin KV store
- Simple to use `/secret` slash command

## Screenshots

![Secret Message](assets/screenshot1.png)

## Installation

1. Download the latest version from the [releases page](https://github.com/mattermost/mattermost-plugin-secrets/releases)
2. Upload the plugin to your Mattermost instance:
   - Navigate to **System Console > Plugins > Plugin Management**
   - Upload the plugin bundle
   - Click "Enable" to activate the plugin

## Usage

### Sending a Secret Message

To send a secret message, use the `/secret` slash command followed by your message:

```
/secret This is a confidential password: p@ssw0rd123
```

This will create a post in the channel with a button that users can click to view the secret message.

### Viewing a Secret Message

1. Click the "View Secret" button on a secret message
2. The message will be displayed only once
3. Optionally, copy the message to clipboard using the copy button
4. Once you navigate away, the message will no longer be accessible to you

## Configuration

The plugin can be configured in the System Console under **Plugins > Secrets Plugin**.

### Available Settings

| Setting | Description | Default |
|---------|-------------|---------|
| Secret Expiry Time | The number of hours after which unviewed secrets will expire | 24 hours |
| Allow Copy to Clipboard | Allow users to copy secret content to clipboard | Enabled |

## Development

### Prerequisites

- Go 1.16 or later
- Node.js 14.x or later
- NPM 6.x or later
- Make

### Building the Plugin

Build the plugin with the following commands:

```bash
# Clone the repository
git clone https://github.com/mattermost/mattermost-plugin-secrets.git
cd mattermost-plugin-secrets

# Build the plugin
make build
```

This will produce a `dist/com.mattermost.secrets-plugin-x.y.z.tar.gz` file that can be uploaded to your Mattermost server.

### Running Tests

Run the tests with:

```bash
make test
```

## License

This repository is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for the full license text.