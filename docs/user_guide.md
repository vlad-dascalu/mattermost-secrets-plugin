# User Guide

This guide explains how to use the Mattermost Secrets Plugin to send and view secret messages that disappear after being viewed once.

## What is a Secret Message?

A secret message is a piece of text that:

- Can only be viewed once by each recipient
- Disappears after being viewed
- Has an expiration time (configurable)
- Is stored securely
- Can be part of a thread
- Supports formatted text (code blocks, multi-line content)

This is ideal for sending sensitive information like passwords, API keys, or other private data that shouldn't persist in chat history.

## Sending a Secret Message

### Using the Slash Command

The easiest way to send a secret message is with the `/secret` slash command:

1. Type `/secret` followed by your message
   ```
   /secret Your temporary password is: P@ssw0rd123!
   ```

2. Press Enter to send the message

3. A message will appear in the channel indicating that you've sent a secret message

### Multi-line Secrets

The plugin supports multi-line secrets for sharing formatted content:

1. Type `/secret` on the first line
2. Press Shift+Enter to start a new line
3. Type your content on subsequent lines

Example:
```
/secret
server: api.example.com
username: admin
password: s3cr3t
api_key: abcd1234efgh5678
```

The formatting will be preserved when the secret is viewed.

### Thread Support

You can send secrets as part of a thread:

1. Reply to an existing message in a thread
2. Use the `/secret` command in your reply
3. The secret will be properly threaded and the viewing interface will appear in the thread context

### Who Can See the Secret?

Anyone in the channel where you send the secret will see that a secret message exists, but only those who click the "View Secret" button will be able to see the actual content of the message.

## Viewing a Secret Message

When someone sends a secret message, you'll see a post with a "View Secret" button:

1. Click the "View Secret" button to view the message content

2. The secret message content will be displayed in a secure container

3. You can copy the content to your clipboard by clicking the copy icon

4. The secret will be marked as viewed for your user

5. Once you navigate away from the conversation, the secret will no longer be visible to you

6. If the secret is part of a thread, the viewing interface will appear in the thread context

## Important Notes

- **View only once**: Each user can only view a secret message once. After viewing, it cannot be accessed again.

- **No record keeping**: Secret messages are not included in message search or export.

- **Expiration**: Unviewed secrets automatically expire after a time period set by your system administrator (default: 60 minutes).

- **Security**: While the plugin secures messages from casual viewing, it's not designed for high-security environments. The messages are stored in the Mattermost database, encrypted according to your server's configuration.

- **Thread Support**: Secrets can be part of threads and maintain proper threading context.

- **Format Preservation**: Multi-line formatting and code blocks are preserved when viewing secrets.

## Use Cases

The Secrets Plugin is ideal for:

- Sharing temporary passwords
- Sending API keys or tokens
- Communicating sensitive personal information
- Sharing confidential URLs or access information
- Sharing formatted code snippets or configuration files
- Thread-based secret discussions

## FAQ

### Can I delete a secret after sending it?

Yes, you can delete the message containing the secret. This will prevent anyone who hasn't already viewed it from accessing it.

### How long do secrets last if not viewed?

By default, secrets expire after 60 minutes if not viewed. Your system administrator can adjust this time period.

### Can team or system admins view secrets?

No, once a secret is created, only the intended recipients can view it, and only once. Not even system administrators can view the content of a secret that has already been viewed.

### Are secrets encrypted?

Secrets are stored in the Mattermost database. The level of encryption depends on your server's configuration. Consult your system administrator for specific security details.

### Can I use secrets in threads?

Yes, the plugin fully supports sending and viewing secrets within threads. The secret viewing interface will appear in the proper thread context.

### Can I share formatted text or code blocks?

Yes, the plugin preserves formatting for multi-line content, including code blocks and structured data. This makes it ideal for sharing configuration files or code snippets securely. 