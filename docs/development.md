# Developer Documentation

This document provides detailed information for developers who want to understand, customize, or contribute to the Mattermost Secrets Plugin.

## Architecture Overview

The Secrets Plugin follows a standard Mattermost plugin architecture with both server-side and webapp components:

### Server-Side Components

The server-side is written in Go and includes these main components:

1. **Plugin Core (`plugin.go`)**: 
   - Implements the Mattermost plugin interface
   - Handles plugin lifecycle events
   - Manages HTTP endpoints
   - Implements slash command handling
   - Handles secret cleanup and expiration

2. **Configuration (`configuration.go`)**: 
   - Manages plugin configuration settings
   - Handles configuration validation
   - Provides type-safe access to settings

3. **Secret Store (`store/secret_store.go`)**: 
   - Provides an interface for storing and retrieving secrets
   - Implements secure storage using Mattermost's KV store
   - Handles secret expiration and cleanup

4. **Models (`models/secret.go`)**: 
   - Defines the data structures used by the plugin
   - Includes validation logic
   - Handles secret viewing tracking

### Webapp Components

The webapp side is written in JavaScript/React and includes:

1. **Index (`index.js`)**: 
   - Main entry point that registers components and actions
   - Sets up Redux store
   - Registers custom post types

2. **SecretPostType (`components/secret_post_type.jsx`)**: 
   - Custom post type rendering for secret messages
   - Handles secret viewing interface
   - Implements copy to clipboard functionality
   - Manages thread context display

3. **Actions (`actions/index.js`)**: 
   - Functions for interacting with the server-side plugin API
   - Handles secret creation and viewing
   - Manages error handling and loading states

4. **Reducers (`reducers/index.js`)**: 
   - Redux reducers for managing state
   - Handles secret viewing status
   - Manages UI state

## Data Flow

1. A user creates a secret message using the `/secret` slash command
2. The server-side plugin handles the command and:
   - Creates a new secret with a unique ID
   - Stores the secret in the KV store
   - Posts a message with a "View Secret" button
   - Handles thread context if present
3. When a user clicks the "View Secret" button:
   - The webapp calls the server API to retrieve the secret
   - The server marks the secret as viewed by that user
   - The webapp displays the secret content to the user
   - Formatting is preserved for multi-line content
   - Once the user navigates away, the secret is no longer displayed

## API Endpoints

The plugin exposes the following REST API endpoints:

### Create Secret

```
POST /plugins/secrets-plugin/api/v1/secrets
```

Request body:
```json
{
  "channel_id": "string",
  "message": "string",
  "root_id": "string"  // Optional, for thread support
}
```

Response:
```json
{
  "id": "string",
  "user_id": "string",
  "channel_id": "string",
  "root_id": "string",
  "viewed_by": ["string"],
  "created_at": 0,
  "expires_at": 0
}
```

### Mark Secret as Viewed

```
POST /plugins/secrets-plugin/api/v1/secrets/viewed
```

Request body:
```json
{
  "secret_id": "string"
}
```

Response: Status 200 OK

### View Secret

```
GET /plugins/secrets-plugin/api/v1/secrets/view?secret_id=string
```

Response:
```json
{
  "id": "string",
  "message": "string",
  "expires_at": 0
}
```

## Secret Storage

Secrets are stored in the Mattermost KV store with the following structure:

```
secret_<id>: {
  "id": "string",
  "user_id": "string",
  "channel_id": "string",
  "root_id": "string",
  "message": "string",
  "viewed_by": ["string"],
  "created_at": 0,
  "expires_at": 0
}
```

Where:
- `id` is a unique identifier for the secret
- `user_id` is the ID of the user who created the secret
- `channel_id` is the channel where the secret was posted
- `root_id` is the ID of the parent post for threaded secrets
- `message` is the content of the secret
- `viewed_by` is a list of user IDs who have viewed the secret
- `created_at` is the time when the secret was created (in milliseconds since epoch)
- `expires_at` is the time when the secret will expire (in milliseconds since epoch)

## Adding New Features

### Adding a New Command

To add a new command, modify the `OnActivate` function in `plugin.go` to register an additional command:

```go
if err := p.API.RegisterCommand(&model.Command{
    Trigger:          "newcommand",
    DisplayName:      "New Command",
    Description:      "Description of the new command",
    AutoComplete:     true,
    AutoCompleteDesc: "Auto-complete description",
    AutoCompleteHint: "[hint]",
}); err != nil {
    return errors.Wrap(err, "failed to register command")
}
```

Then implement the command handler by extending the `ExecuteCommand` function.

### Adding a New Configuration Option

To add a new configuration option:

1. Update the `configuration` struct in `configuration.go`
2. Update the `settings` array in the `settings_schema` section of `plugin.json`
3. Access the configuration value using `p.getConfiguration()` in your code

### Adding Thread Support to a New Feature

To add thread support to a new feature:

1. Include `root_id` in your data structures
2. Pass the `root_id` when creating posts
3. Handle thread context in the UI components
4. Update the API endpoints to support thread-related parameters

## Testing

### Server-Side Testing

Server-side tests use the standard Go testing package and the Mattermost plugin test helpers. To run the tests:

```bash
cd server
go test ./...
```

### Webapp Testing

Webapp tests use Jest and React Testing Library. To run the tests:

```bash
cd webapp
npm test
```

### Testing Thread Support

When testing thread support:

1. Create test cases for threaded and non-threaded secrets
2. Verify proper thread context preservation
3. Test secret viewing in thread context
4. Verify thread UI integration

## Troubleshooting

### Common Issues

1. **Plugin fails to activate**: Check the Mattermost server logs for errors.
2. **Slash command not registered**: Ensure the command is properly registered in `OnActivate`.
3. **KV store errors**: Check for issues with the Mattermost KV store permissions.
4. **Thread context issues**: Verify proper handling of `root_id` in posts and API calls.
5. **Format preservation**: Check if multi-line content is properly escaped and preserved.

### Debugging

Enable debug logging in the Mattermost System Console to see more detailed logs from the plugin.

### Performance Considerations

1. **Secret Cleanup**: The plugin runs periodic cleanup of expired secrets to prevent database bloat.
2. **Thread Loading**: Consider lazy loading of thread content for better performance.
3. **KV Store Usage**: Monitor KV store usage as it can impact performance with many secrets. 