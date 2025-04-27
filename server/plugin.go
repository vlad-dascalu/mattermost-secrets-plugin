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

	"github.com/mattermost/mattermost-plugin-secrets/server/models"
	"github.com/mattermost/mattermost-plugin-secrets/server/store"
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

	// Create the secret
	secret, err := p.createSecret(userID, req.ChannelID, req.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	// Mark the secret as viewed by this user
	if err := p.markSecretAsViewed(req.SecretID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleViewSecret handles requests when a user clicks the View Secret button
func (p *Plugin) handleViewSecret(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For button actions, we need to parse the form values
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Get the secret ID from the query parameters
	secretID := r.URL.Query().Get("secret_id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Get the secret
	secret, err := p.secretStore.GetSecret(secretID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get secret: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if secret == nil {
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	// Check if the secret has expired
	if secret.ExpiresAt > 0 && secret.ExpiresAt < models.GetMillis() {
		// Secret has expired, delete it
		if err := p.secretStore.DeleteSecret(secretID); err != nil {
			p.API.LogError("Failed to delete expired secret", "secret_id", secretID, "error", err.Error())
		}

		// Return an error response
		http.Error(w, "Secret has expired", http.StatusGone)
		return
	}

	// Check if the user has already viewed this secret
	userViewed := false
	for _, id := range secret.ViewedBy {
		if id == userID {
			userViewed = true
			break
		}
	}

	// If the user hasn't viewed it yet, mark it as viewed
	if !userViewed {
		if err := p.markSecretAsViewed(secretID, userID); err != nil {
			http.Error(w, fmt.Sprintf("Failed to mark secret as viewed: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}

	// Get the user who created the secret for display purposes
	user, appErr := p.API.GetUser(secret.UserID)
	var username string
	if appErr != nil {
		p.API.LogError("Failed to get user", "error", appErr.Error())
		username = "Unknown User"
	} else {
		username = user.Username
	}

	// Check if this is a "close" action
	closeAction := r.URL.Query().Get("action") == "close"
	if closeAction {
		// Extract the post ID from the context
		postID := r.URL.Query().Get("post_id")
		if postID == "" {
			postID = "unknown_post_id" // Fallback if post_id is missing
		}

		// Create a response that replaces the message for this user with a simple acknowledgment
		update := &model.PostActionIntegrationResponse{
			Update: &model.Post{
				Id: postID,
				Props: map[string]interface{}{
					"attachments": []*model.SlackAttachment{
						{
							Title: "Secret Message",
							Text:  "This secret message has been closed.",
							Color: "#DDDDDD",
						},
					},
				},
			},
			EphemeralText: "You've closed this secret message.",
		}

		// Return the update that replaces the message
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(update); err != nil {
			p.API.LogError("Failed to write JSON response", "error", err.Error())
		}
		return
	}

	// Construct actions for the response
	actions := []*model.PostAction{
		{
			Id:   model.NewId(),
			Name: "Close",
			Type: "button",
			Integration: &model.PostActionIntegration{
				URL: fmt.Sprintf("/plugins/com.mattermost.secrets-plugin/api/v1/secrets/view?secret_id=%s&action=close&post_id=${post.id}", secret.ID),
			},
		},
	}

	// Construct a proper update to show the secret
	update := &model.PostActionIntegrationResponse{
		Update: &model.Post{
			Props: map[string]interface{}{
				"attachments": []*model.SlackAttachment{
					{
						Title: "Secret Message",
						Text:  fmt.Sprintf("**From @%s:**\n\n```\n%s\n```", username, secret.Message),
						Fields: []*model.SlackAttachmentField{
							{
								Title: "Status",
								Value: "This message can only be viewed once per person. It will be automatically deleted when everyone in the channel has viewed it or when it expires.",
								Short: false,
							},
						},
						Actions: actions,
					},
				},
			},
		},
	}

	// Check if all users have viewed the secret
	// For channels with many members, we'll consider it fully viewed after 75% of members have seen it
	go func() {
		// Get channel members
		var memberCount int

		// Check if this is a DM/GM channel
		channel, appErr := p.API.GetChannel(secret.ChannelID)
		if appErr != nil {
			p.API.LogError("Failed to get channel", "channel_id", secret.ChannelID, "error", appErr.Error())
		} else {
			// Check channel type
			if channel.Type == model.ChannelTypeDirect || channel.Type == model.ChannelTypeGroup {
				// For DMs, just count the channel members
				stats, appErr := p.API.GetChannelStats(secret.ChannelID)
				if appErr != nil {
					p.API.LogError("Failed to get channel stats", "channel_id", secret.ChannelID, "error", appErr.Error())
				} else {
					memberCount = int(stats.MemberCount)
				}
			} else {
				// For regular channels, get total member count
				stats, appErr := p.API.GetChannelStats(secret.ChannelID)
				if appErr != nil {
					p.API.LogError("Failed to get channel stats", "channel_id", secret.ChannelID, "error", appErr.Error())
				} else {
					memberCount = int(stats.MemberCount)
				}
			}
		}

		// No point keeping track if we couldn't determine membership
		if memberCount == 0 {
			memberCount = 10 // Default to a reasonable number
		}

		// Calculate threshold for considering the secret fully viewed
		// Always require 100% of members to view the secret
		viewThreshold := memberCount

		// Check if we've reached the threshold
		if len(secret.ViewedBy) >= viewThreshold {
			// All members have viewed the secret, clean it up
			if err := p.secretStore.DeleteSecret(secret.ID); err != nil {
				p.API.LogError("Failed to delete viewed secret", "secret_id", secret.ID, "error", err.Error())
			}

			// Try to delete the post too
			posts, appErr := p.API.GetPostsForChannel(secret.ChannelID, 0, 100)
			if appErr != nil {
				p.API.LogError("Failed to get posts for channel", "channel_id", secret.ChannelID, "error", appErr.Error())
				return
			}

			// Try to find the post containing this secret
			for _, post := range posts.Posts {
				if attachments, ok := post.Props["attachments"].([]interface{}); ok {
					for _, attachment := range attachments {
						if attach, ok := attachment.(map[string]interface{}); ok {
							if actions, ok := attach["actions"].([]interface{}); ok {
								for _, action := range actions {
									if act, ok := action.(map[string]interface{}); ok {
										if integration, ok := act["integration"].(map[string]interface{}); ok {
											url, ok := integration["url"].(string)
											if ok && strings.Contains(url, secret.ID) {
												// Found the post, delete it after a delay
												time.AfterFunc(5*time.Second, func() {
													if appErr := p.API.DeletePost(post.Id); appErr != nil {
														p.API.LogError("Failed to delete post for viewed secret", "post_id", post.Id, "error", appErr.Error())
													}
												})
												break
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
	}()

	// Return the dialog that shows the secret
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(update); err != nil {
		p.API.LogError("Failed to write JSON response", "error", err.Error())
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

	// Create a response that replaces the message for this user with a simple acknowledgment
	response := &model.PostActionIntegrationResponse{
		Update: &model.Post{
			Id: postID,
			Props: map[string]interface{}{
				"attachments": []*model.SlackAttachment{
					{
						Title: "Secret Message",
						Text:  "This secret message has been closed.",
						Color: "#DDDDDD",
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
			p.API.LogDebug("Deleting expired secret", "secret_id", secret.ID)

			// Delete the post if we can find it
			posts, appErr := p.API.GetPostsForChannel(secret.ChannelID, 0, 100)
			if appErr != nil {
				p.API.LogError("Failed to get posts for channel", "channel_id", secret.ChannelID, "error", appErr.Error())
			} else {
				// Try to find the post containing this secret
				for _, post := range posts.Posts {
					if attachments, ok := post.Props["attachments"].([]interface{}); ok {
						for _, attachment := range attachments {
							if attach, ok := attachment.(map[string]interface{}); ok {
								if actions, ok := attach["actions"].([]interface{}); ok {
									for _, action := range actions {
										if act, ok := action.(map[string]interface{}); ok {
											if integration, ok := act["integration"].(map[string]interface{}); ok {
												url, ok := integration["url"].(string)
												if ok && strings.Contains(url, secret.ID) {
													// Found the post, delete it
													if appErr := p.API.DeletePost(post.Id); appErr != nil {
														p.API.LogError("Failed to delete post for expired secret", "post_id", post.Id, "error", appErr.Error())
													}
													break
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

			// Delete the secret
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
	secret, err := p.createSecret(args.UserId, args.ChannelId, message)
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

	// Create the post with the secret placeholder
	post := &model.Post{
		UserId:    p.botID,
		ChannelId: args.ChannelId,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{
				{
					Title: "Secret Message",
					Text:  fmt.Sprintf("@%s has sent a secret message. Click to view it once.", user.Username),
					Actions: []*model.PostAction{
						{
							Id:   model.NewId(),
							Name: "View Secret",
							Type: "button",
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("/plugins/com.mattermost.secrets-plugin/api/v1/secrets/view?secret_id=%s", secret.ID),
							},
						},
					},
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
func (p *Plugin) createSecret(userID, channelID, message string) (*models.Secret, error) {
	// Create a new secret
	secret := &models.Secret{
		ID:        model.NewId(),
		UserID:    userID,
		ChannelID: channelID,
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
func (p *Plugin) markSecretAsViewed(secretID, userID string) error {
	// Get the secret
	secret, err := p.secretStore.GetSecret(secretID)
	if err != nil {
		return errors.Wrap(err, "failed to get secret")
	}

	if secret == nil {
		return errors.New("secret not found")
	}

	// Check if the user has already viewed the secret
	for _, id := range secret.ViewedBy {
		if id == userID {
			return nil // User has already viewed this secret
		}
	}

	// Mark as viewed by this user
	secret.ViewedBy = append(secret.ViewedBy, userID)

	// Save the updated secret
	if err := p.secretStore.SaveSecret(secret); err != nil {
		return errors.Wrap(err, "failed to update secret")
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

// This function needs to be defined for our plugin to be started
func main() {
	plugin.ClientMain(&Plugin{})
}
