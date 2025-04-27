package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

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

// OnActivate is invoked when the plugin is activated
func (p *Plugin) OnActivate() error {
	// Initialize the secret store
	p.secretStore = store.NewKVSecretStore(p.API)

	// Register the bot account
	bot := &model.Bot{
		Username:    "secrets-bot",
		DisplayName: "Secrets Bot",
		Description: "A bot account for the Secrets plugin",
	}

	botUser, err := p.API.CreateBot(bot)
	if err != nil {
		return errors.Wrap(err, "failed to create bot account")
	}
	p.botID = botUser.UserId

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

	return nil
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

	_, err = p.API.CreatePost(post)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error creating post: %s", err.Error()),
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
		ExpiresAt: models.GetMillis() + (int64(p.getConfiguration().SecretExpiryTime) * 60 * 60 * 1000), // Convert hours to milliseconds
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
