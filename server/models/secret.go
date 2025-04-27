package models

// Secret represents a secret message that can only be viewed once by each user
type Secret struct {
	// ID is the unique identifier for the secret
	ID string `json:"id"`

	// UserID is the ID of the user who created the secret
	UserID string `json:"user_id"`

	// ChannelID is the channel where the secret was posted
	ChannelID string `json:"channel_id"`

	// RootId is the ID of the parent post if the secret is in a thread
	RootId string `json:"root_id"`

	// Message is the content of the secret message
	Message string `json:"message"`

	// ViewedBy is a list of user IDs who have viewed this secret
	ViewedBy []string `json:"viewed_by"`

	// CreatedAt is the time when the secret was created (in milliseconds since epoch)
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is the time when the secret will expire (in milliseconds since epoch)
	ExpiresAt int64 `json:"expires_at"`
}

// SecretRequest is used when creating a new secret via the API
type SecretRequest struct {
	// ChannelID is the channel where the secret should be posted
	ChannelID string `json:"channel_id"`

	// RootId is the ID of the parent post if this secret should be in a thread
	RootId string `json:"root_id"`

	// Message is the content of the secret message
	Message string `json:"message"`
}

// SecretViewedRequest is used when marking a secret as viewed via the API
type SecretViewedRequest struct {
	// SecretID is the ID of the secret being viewed
	SecretID string `json:"secret_id"`
}

// SecretResponse is sent when a user views a secret
type SecretResponse struct {
	// Message is the content of the secret message
	Message string `json:"message"`

	// AllowCopy indicates whether the user is allowed to copy the secret
	AllowCopy bool `json:"allow_copy"`
}
