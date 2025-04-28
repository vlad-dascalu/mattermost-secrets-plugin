package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"

	"github.com/vlad-dascalu/mattermost-secrets-plugin/server/models"
	"github.com/vlad-dascalu/mattermost-secrets-plugin/server/store"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// BotID is the ID of the bot user for this plugin
	botID string

	// secretStore manages persistence and retrieval of secrets
	secretStore store.SecretStore
}

// ServeHTTP demonstrates a plugin that handles HTTP requests.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/secrets":
		p.handleSecret(w, r)
	case "/api/v1/secrets/viewed":
		p.handleSecretViewed(w, r)
	case "/api/v1/secrets/view":
		p.handleViewSecret(w, r)
	case "/api/v1/secrets/close":
		p.handleCloseSecret(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleSecret handles requests for creating a new secret message
func (p *Plugin) handleSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only authenticated users can create secrets
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req models.SecretRequest
	if err := p.parseJSONBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ChannelID == "" || req.Message == "" {
		http.Error(w, "channelId and message are required", http.StatusBadRequest)
		return
	}

	// If RootId is provided, verify it exists
	if req.RootId != "" {
		_, appErr := p.API.GetPost(req.RootId)
		if appErr != nil {
			http.Error(w, fmt.Sprintf("Invalid rootId: %s", appErr.Error()), http.StatusBadRequest)
			return
		}
	}

	// Create the secret
	secret, err := p.createSecret(userID, req.ChannelID, req.Message, req.RootId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the user who created the secret
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("Failed to get user", "error", appErr.Error())
		p.writeJSON(w, secret)
		return
	}

	// Create the post with the custom post type
	post := &model.Post{
		UserId:    p.botID,
		ChannelId: req.ChannelID,
		RootId:    req.RootId,
		Type:      "custom_secret",
		Props: map[string]interface{}{
			"secret_id": secret.ID,
			"attachments": []*model.SlackAttachment{
				{
					Title: "Secret Message",
					Text:  fmt.Sprintf("@%s has sent a secret message.", user.Username),
				},
			},
		},
	}

	_, postErr := p.API.CreatePost(post)
	if postErr != nil {
		p.API.LogError("Failed to create post", "error", postErr.Error())
	}

	p.writeJSON(w, secret)
}

// handleSecretViewed handles requests when a user views a secret
func (p *Plugin) handleSecretViewed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only authenticated users can view secrets
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req models.SecretViewedRequest
	if err := p.parseJSONBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.SecretID == "" {
		http.Error(w, "secretId is required", http.StatusBadRequest)
		return
	}

	// First retrieve the secret object
	secret, err := p.secretStore.GetSecret(req.SecretID)
	if err != nil {
		p.API.LogError("Failed to get secret", "error", err.Error())
		http.Error(w, "Failed to get secret", http.StatusInternalServerError)
		return
	}

	if secret == nil {
		p.API.LogWarn("Secret not found", "secret_id", req.SecretID)
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	// Mark the secret as viewed by this user
	if err := p.markSecretAsViewed(secret, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleViewSecret handles requests when a user clicks the View Secret button
func (p *Plugin) handleViewSecret(w http.ResponseWriter, r *http.Request) {
	secretID := r.URL.Query().Get("secret_id")
	if secretID == "" {
		p.API.LogWarn("No secret ID provided to view secret endpoint")
		http.Error(w, "No secret ID provided", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		p.API.LogWarn("No user ID found in request to view secret")
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	secret, err := p.secretStore.GetSecret(secretID)
	if err != nil {
		p.API.LogError("Failed to get secret", "error", err.Error())
		http.Error(w, "Failed to get secret", http.StatusInternalServerError)
		return
	}

	if secret == nil {
		p.API.LogWarn("Secret not found", "secret_id", secretID)
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	// Check if the secret has expired
	currentTime := models.GetMillis()
	if secret.ExpiresAt <= currentTime {
		p.API.LogDebug("Attempted to view expired secret", "secret_id", secretID, "user_id", userID)

		// Send an ephemeral post to the user indicating the secret has expired
		expiredPost := &model.Post{
			UserId:    p.botID,
			ChannelId: secret.ChannelID,
			Message:   "**This secret has expired and is no longer available.**",
			RootId:    secret.RootId, // Include the RootId to make the ephemeral message appear in the thread
		}
		p.API.SendEphemeralPost(userID, expiredPost)

		// Update the post to show it's expired
		p.updatePostForExpiredSecret(secret)

		// Send a response for the integration
		response := &model.PostActionIntegrationResponse{
			EphemeralText: "Secret has expired.",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			p.API.LogError("Failed to encode response", "error", err.Error())
		}
		return
	}

	// Add this user to the ViewedBy list
	err = p.markSecretAsViewed(secret, userID)
	if err != nil {
		p.API.LogError("Failed to mark secret as viewed", "error", err.Error())
		http.Error(w, "Failed to mark secret as viewed", http.StatusInternalServerError)
		return
	}

	// Send an ephemeral post directly to the user
	ephemeralPost := &model.Post{
		UserId:    p.botID,
		ChannelId: secret.ChannelID,
		Message:   "**Secret Message**:\n```\n" + secret.Message + "\n```",
		RootId:    secret.RootId, // Include the RootId to make the ephemeral message appear in the thread
	}

	p.API.SendEphemeralPost(userID, ephemeralPost)

	p.API.LogDebug("Sending ephemeral message with secret content",
		"secret_id", secretID,
		"user_id", userID,
		"channel_id", secret.ChannelID)

	// Also send a response for the integration
	response := &model.PostActionIntegrationResponse{}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.API.LogError("Failed to encode response", "error", err.Error())
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}
}

// maybeCleanupSecret checks if a secret should be cleaned up and handles it if needed
func (p *Plugin) maybeCleanupSecret(secret *models.Secret) {
	// Get channel members
	var memberCount int

	// Check if this is a DM/GM channel
	channel, appErr := p.API.GetChannel(secret.ChannelID)
	if appErr != nil {
		p.API.LogError("Failed to get channel", "channel_id", secret.ChannelID, "error", appErr.Error())
		return
	}

	// Check channel type
	if channel.Type == model.ChannelTypeDirect || channel.Type == model.ChannelTypeGroup {
		// For DMs, just count the channel members
		stats, appErr := p.API.GetChannelStats(secret.ChannelID)
		if appErr != nil {
			p.API.LogError("Failed to get channel stats", "channel_id", secret.ChannelID, "error", appErr.Error())
			return
		}
		memberCount = int(stats.MemberCount)
	} else {
		// For regular channels, get total member count
		stats, appErr := p.API.GetChannelStats(secret.ChannelID)
		if appErr != nil {
			p.API.LogError("Failed to get channel stats", "channel_id", secret.ChannelID, "error", appErr.Error())
			return
		}
		memberCount = int(stats.MemberCount)
	}

	// No point keeping track if we couldn't determine membership
	if memberCount == 0 {
		memberCount = 10 // Default to a reasonable number
	}

	// Calculate threshold for considering the secret fully viewed
	viewThreshold := memberCount

	p.API.LogDebug("Checking secret cleanup status",
		"secret_id", secret.ID,
		"viewed_by_count", len(secret.ViewedBy),
		"member_count", memberCount,
		"view_threshold", viewThreshold)

	// Check if we've reached the threshold
	if len(secret.ViewedBy) >= viewThreshold {
		p.API.LogDebug("Secret has been viewed by all members, cleaning up", "secret_id", secret.ID)

		// All members have viewed the secret, clean it up
		if err := p.secretStore.DeleteSecret(secret.ID); err != nil {
			p.API.LogError("Failed to delete viewed secret", "secret_id", secret.ID, "error", err.Error())
		}

		// Try to find and delete the post too
		p.deletePostBySecretID(secret.ID, secret.ChannelID)
	}
}

// deletePostBySecretID finds and deletes a post associated with a secret
func (p *Plugin) deletePostBySecretID(secretID, channelID string) {
	posts, appErr := p.API.GetPostsForChannel(channelID, 0, 100)
	if appErr != nil {
		p.API.LogError("Failed to get posts for channel", "channel_id", channelID, "error", appErr.Error())
		return
	}

	for _, post := range posts.Posts {
		secretIDFromPost, ok := post.Props["secret_id"].(string)
		if ok && secretIDFromPost == secretID {
			// Found the post with our secret ID
			p.API.LogDebug("Found post for secret, deleting", "post_id", post.Id, "secret_id", secretID)
			time.AfterFunc(5*time.Second, func() {
				if appErr := p.API.DeletePost(post.Id); appErr != nil {
					p.API.LogError("Failed to delete post for viewed secret", "post_id", post.Id, "error", appErr.Error())
				}
			})
			return
		}

		// Check for the old style post with button actions
		if attachments, ok := post.Props["attachments"].([]interface{}); ok {
			for _, attachment := range attachments {
				if attach, ok := attachment.(map[string]interface{}); ok {
					if actions, ok := attach["actions"].([]interface{}); ok {
						for _, action := range actions {
							if act, ok := action.(map[string]interface{}); ok {
								if integration, ok := act["integration"].(map[string]interface{}); ok {
									url, ok := integration["url"].(string)
									if ok && strings.Contains(url, secretID) {
										// Found the post, delete it after a delay
										p.API.LogDebug("Found post with action for secret, deleting", "post_id", post.Id, "secret_id", secretID)
										time.AfterFunc(5*time.Second, func() {
											if appErr := p.API.DeletePost(post.Id); appErr != nil {
												p.API.LogError("Failed to delete post for viewed secret", "post_id", post.Id, "error", appErr.Error())
											}
										})
										return
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// handleCloseSecret handles closing a secret for a specific user
func (p *Plugin) handleCloseSecret(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the secret ID from the query parameters
	secretID := r.URL.Query().Get("secret_id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Extract the post ID from the context
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		postID = "unknown_post_id" // Fallback if post_id is missing
	}

	// Check if the secret exists and if it has expired
	secret, err := p.secretStore.GetSecret(secretID)
	if err != nil {
		p.API.LogError("Failed to get secret", "error", err.Error())
		http.Error(w, "Failed to get secret", http.StatusInternalServerError)
		return
	}

	// Prepare message based on whether the secret exists and is expired
	message := "This secret message has been closed."
	color := "#DDDDDD"

	if secret == nil {
		message = "This secret message is no longer available."
	} else if secret.ExpiresAt <= models.GetMillis() {
		message = "This secret message has expired and is no longer available."
	}

	// Create a response that replaces the message for this user with a simple acknowledgment
	response := &model.PostActionIntegrationResponse{
		Update: &model.Post{
			Id: postID,
			Props: map[string]interface{}{
				"attachments": []*model.SlackAttachment{
					{
						Title: "Secret Message",
						Text:  message,
						Color: color,
					},
				},
			},
		},
		EphemeralText: "You've closed this secret message.",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.API.LogError("Failed to write JSON response", "error", err.Error())
	}
}

// OnActivate is invoked when the plugin is activated
func (p *Plugin) OnActivate() error {
	// Initialize the secret store
	p.secretStore = store.NewKVSecretStore(p.API)

	// Define bot user
	botUsername := "secrets-bot"
	botDisplayName := "Secrets Bot"
	botDescription := "A bot account for the Secrets plugin"

	// Try to find the existing bot
	bot, err := p.API.GetUserByUsername(botUsername)
	if err != nil {
		// Bot doesn't exist yet, create it
		botUser, createErr := p.API.CreateBot(&model.Bot{
			Username:    botUsername,
			DisplayName: botDisplayName,
			Description: botDescription,
		})
		if createErr != nil {
			// If we can't create because it exists, try to get it again
			if strings.Contains(createErr.Error(), "store.sql_bot.save.exists.app_error") {
				existingBot, getErr := p.API.GetUserByUsername(botUsername)
				if getErr != nil {
					return errors.Wrap(getErr, "failed to get existing bot")
				}
				p.botID = existingBot.Id
			} else {
				return errors.Wrap(createErr, "failed to create bot account")
			}
		} else {
			p.botID = botUser.UserId
		}
	} else {
		// Bot exists, use it
		p.botID = bot.Id
	}

	// Register the slash command
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          "secret",
		DisplayName:      "Secret Message",
		Description:      "Send a secret message that disappears after being viewed",
		AutoComplete:     true,
		AutoCompleteDesc: "Create a secret message",
		AutoCompleteHint: "[message]",
	}); err != nil {
		return errors.Wrap(err, "failed to register command")
	}

	// Start a routine to clean up expired secrets
	go p.periodicCleanup()

	return nil
}

// periodicCleanup periodically checks for and deletes expired secrets
func (p *Plugin) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanupExpiredSecrets()
		}
	}
}

// cleanupExpiredSecrets finds and removes expired secrets
func (p *Plugin) cleanupExpiredSecrets() {
	p.API.LogDebug("Checking for expired secrets")

	// Get all secrets
	secrets, err := p.secretStore.GetAllSecrets()
	if err != nil {
		p.API.LogError("Failed to get secrets for cleanup", "error", err.Error())
		return
	}

	currentTime := models.GetMillis()
	for _, secret := range secrets {
		// Check if the secret has expired
		if secret.ExpiresAt <= currentTime {
			p.API.LogDebug("Found expired secret during cleanup", "secret_id", secret.ID)

			// First update the post to show it's expired
			p.updatePostForExpiredSecret(secret)

			// Then delete the secret
			if err := p.secretStore.DeleteSecret(secret.ID); err != nil {
				p.API.LogError("Failed to delete expired secret", "secret_id", secret.ID, "error", err.Error())
			}
		}
	}
}

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database. We'll check if it's a secret message command.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	// We'll handle secret slash commands in ExecuteCommand
	return post, ""
}

// ExecuteCommand handles the /secret slash command
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Skip the command name (/secret)
	message := args.Command[len("/secret"):]

	// Trim whitespace
	message = strings.TrimSpace(message)

	if message == "" {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Please provide a message to be kept secret.",
		}, nil
	}

	// Create the secret
	secret, err := p.createSecret(args.UserId, args.ChannelId, message, args.RootId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error creating secret: %s", err.Error()),
		}, nil
	}

	// Get the user who created the secret
	user, appErr := p.API.GetUser(args.UserId)
	if appErr != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error getting user: %s", appErr.Error()),
		}, nil
	}

	// Create the post with the custom post type
	post := &model.Post{
		UserId:    p.botID,
		ChannelId: args.ChannelId,
		RootId:    args.RootId,
		Type:      "custom_secret",
		Props: map[string]interface{}{
			"secret_id": secret.ID,
			"attachments": []*model.SlackAttachment{
				{
					Title: "Secret Message",
					Text:  fmt.Sprintf("@%s has sent a secret message.", user.Username),
				},
			},
		},
	}

	var postErr *model.AppError
	_, postErr = p.API.CreatePost(post)
	if postErr != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error creating post: %s", postErr.Error()),
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Secret message created successfully!",
	}, nil
}

// createSecret creates a new secret message
func (p *Plugin) createSecret(userID, channelID, message string, rootID string) (*models.Secret, error) {
	// Create a new secret
	secret := &models.Secret{
		ID:        model.NewId(),
		UserID:    userID,
		ChannelID: channelID,
		RootId:    rootID,
		Message:   message,
		ViewedBy:  []string{},
		CreatedAt: models.GetMillis(),
		ExpiresAt: models.GetMillis() + (int64(p.getConfiguration().SecretExpiryTime) * 60 * 1000), // Convert minutes to milliseconds
	}

	// Save the secret
	if err := p.secretStore.SaveSecret(secret); err != nil {
		return nil, errors.Wrap(err, "failed to save secret")
	}

	return secret, nil
}

// markSecretAsViewed marks a secret as viewed by a user
func (p *Plugin) markSecretAsViewed(secret *models.Secret, userID string) error {
	// Check if secret is nil
	if secret == nil {
		return errors.New("secret not found")
	}

	p.API.LogDebug("Marking secret as viewed", "secret_id", secret.ID, "user_id", userID)

	// Check if the user has already viewed the secret
	userAlreadyViewed := false
	for _, id := range secret.ViewedBy {
		if id == userID {
			userAlreadyViewed = true
			p.API.LogDebug("User has already viewed this secret", "user_id", userID, "secret_id", secret.ID)
			return nil // User has already viewed this secret
		}
	}

	if !userAlreadyViewed {
		// Mark as viewed by this user
		p.API.LogDebug("Adding user to ViewedBy list", "user_id", userID, "secret_id", secret.ID, "current_viewed_count", len(secret.ViewedBy))
		secret.ViewedBy = append(secret.ViewedBy, userID)

		// Save the updated secret
		if err := p.secretStore.SaveSecret(secret); err != nil {
			p.API.LogError("Failed to save secret after marking as viewed", "secret_id", secret.ID, "error", err.Error())
			return errors.Wrap(err, "failed to update secret")
		}

		p.API.LogDebug("Successfully marked secret as viewed", "user_id", userID, "secret_id", secret.ID, "viewed_count", len(secret.ViewedBy))
	}

	return nil
}

// Helper to parse JSON body
func (p *Plugin) parseJSONBody(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.Wrap(err, "failed to decode request body")
	}
	return nil
}

// Helper to write JSON response
func (p *Plugin) writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		p.API.LogError("Failed to write JSON response", "error", err.Error())
	}
}

// updatePostForExpiredSecret updates the UI of a post containing an expired secret
func (p *Plugin) updatePostForExpiredSecret(secret *models.Secret) {
	posts, appErr := p.API.GetPostsForChannel(secret.ChannelID, 0, 100)
	if appErr != nil {
		p.API.LogError("Failed to get posts for channel", "channel_id", secret.ChannelID, "error", appErr.Error())
		return
	}

	for _, post := range posts.Posts {
		secretIDFromPost, ok := post.Props["secret_id"].(string)
		if ok && secretIDFromPost == secret.ID {
			// Found the post with our secret ID
			p.API.LogDebug("Found post for expired secret, updating UI", "post_id", post.Id, "secret_id", secret.ID)

			// Update the post to indicate the secret has expired
			updatedPost := post.Clone()
			updatedPost.Props["expired"] = true

			// Update attachments to show expiration message instead of the original message
			if attachments, ok := updatedPost.Props["attachments"].([]interface{}); ok && len(attachments) > 0 {
				if attach, ok := attachments[0].(map[string]interface{}); ok {
					attach["text"] = "This secret message has expired and is no longer available."
					attach["color"] = "#DDDDDD" // Gray color to indicate expiration

					// Remove any actions that might have been present
					delete(attach, "actions")

					attachments[0] = attach
					updatedPost.Props["attachments"] = attachments
				}
			}

			if _, err := p.API.UpdatePost(updatedPost); err != nil {
				p.API.LogError("Failed to update post for expired secret", "post_id", post.Id, "error", err.Error())
			}
			return
		}

		// Check for the old style post with button actions
		if attachments, ok := post.Props["attachments"].([]interface{}); ok {
			for i, attachment := range attachments {
				if attach, ok := attachment.(map[string]interface{}); ok {
					if actions, ok := attach["actions"].([]interface{}); ok {
						for _, action := range actions {
							if act, ok := action.(map[string]interface{}); ok {
								if integration, ok := act["integration"].(map[string]interface{}); ok {
									url, ok := integration["url"].(string)
									if ok && strings.Contains(url, secret.ID) {
										// Found the post, update it
										p.API.LogDebug("Found post with action for expired secret, updating", "post_id", post.Id, "secret_id", secret.ID)

										updatedPost := post.Clone()
										updatedAttachments := make([]interface{}, len(attachments))
										copy(updatedAttachments, attachments)

										// Update this specific attachment
										updatedAttach := map[string]interface{}{}
										for k, v := range attach {
											updatedAttach[k] = v
										}

										// Remove actions and update text
										delete(updatedAttach, "actions")
										updatedAttach["text"] = "This secret message has expired and is no longer available."
										updatedAttach["color"] = "#DDDDDD" // Gray color

										updatedAttachments[i] = updatedAttach
										updatedPost.Props["attachments"] = updatedAttachments
										updatedPost.Props["expired"] = true

										if _, err := p.API.UpdatePost(updatedPost); err != nil {
											p.API.LogError("Failed to update post for expired secret", "post_id", post.Id, "error", err.Error())
										}
										return
									}
								}
							}
						}
					}
				}
			}
		}
	}

	p.API.LogDebug("Could not find post for expired secret to update", "secret_id", secret.ID)
}

// This function needs to be defined for our plugin to be started
func main() {
	plugin.ClientMain(&Plugin{})
}
