package models

import (
	"encoding/json"
	"testing"
)

func TestSecretJSON(t *testing.T) {
	// Test marshaling and unmarshaling of Secret
	secret := &Secret{
		ID:        "test-id",
		UserID:    "user-id",
		ChannelID: "channel-id",
		RootId:    "root-id",
		Message:   "test message",
		ViewedBy:  []string{"user1", "user2"},
		CreatedAt: GetMillis(),
		ExpiresAt: GetMillis() + 3600000, // 1 hour from now
	}

	// Marshal to JSON
	data, err := json.Marshal(secret)
	if err != nil {
		t.Errorf("Failed to marshal Secret: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled Secret
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal Secret: %v", err)
	}

	// Compare fields
	if secret.ID != unmarshaled.ID {
		t.Errorf("ID mismatch: got %v, want %v", unmarshaled.ID, secret.ID)
	}
	if secret.UserID != unmarshaled.UserID {
		t.Errorf("UserID mismatch: got %v, want %v", unmarshaled.UserID, secret.UserID)
	}
	if secret.ChannelID != unmarshaled.ChannelID {
		t.Errorf("ChannelID mismatch: got %v, want %v", unmarshaled.ChannelID, secret.ChannelID)
	}
	if secret.RootId != unmarshaled.RootId {
		t.Errorf("RootId mismatch: got %v, want %v", unmarshaled.RootId, secret.RootId)
	}
	if secret.Message != unmarshaled.Message {
		t.Errorf("Message mismatch: got %v, want %v", unmarshaled.Message, secret.Message)
	}
	if len(secret.ViewedBy) != len(unmarshaled.ViewedBy) {
		t.Errorf("ViewedBy length mismatch: got %v, want %v", len(unmarshaled.ViewedBy), len(secret.ViewedBy))
	}
	if secret.CreatedAt != unmarshaled.CreatedAt {
		t.Errorf("CreatedAt mismatch: got %v, want %v", unmarshaled.CreatedAt, secret.CreatedAt)
	}
	if secret.ExpiresAt != unmarshaled.ExpiresAt {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", unmarshaled.ExpiresAt, secret.ExpiresAt)
	}
}

func TestSecretRequestJSON(t *testing.T) {
	// Test marshaling and unmarshaling of SecretRequest
	request := &SecretRequest{
		ChannelID: "channel-id",
		RootId:    "root-id",
		Message:   "test message",
	}

	// Marshal to JSON
	data, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal SecretRequest: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled SecretRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SecretRequest: %v", err)
	}

	// Compare fields
	if request.ChannelID != unmarshaled.ChannelID {
		t.Errorf("ChannelID mismatch: got %v, want %v", unmarshaled.ChannelID, request.ChannelID)
	}
	if request.RootId != unmarshaled.RootId {
		t.Errorf("RootId mismatch: got %v, want %v", unmarshaled.RootId, request.RootId)
	}
	if request.Message != unmarshaled.Message {
		t.Errorf("Message mismatch: got %v, want %v", unmarshaled.Message, request.Message)
	}
}

func TestSecretViewedRequestJSON(t *testing.T) {
	// Test marshaling and unmarshaling of SecretViewedRequest
	request := &SecretViewedRequest{
		SecretID: "secret-id",
	}

	// Marshal to JSON
	data, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal SecretViewedRequest: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled SecretViewedRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SecretViewedRequest: %v", err)
	}

	// Compare fields
	if request.SecretID != unmarshaled.SecretID {
		t.Errorf("SecretID mismatch: got %v, want %v", unmarshaled.SecretID, request.SecretID)
	}
}

func TestSecretResponseJSON(t *testing.T) {
	// Test marshaling and unmarshaling of SecretResponse
	response := &SecretResponse{
		Message:   "test message",
		AllowCopy: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal SecretResponse: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled SecretResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SecretResponse: %v", err)
	}

	// Compare fields
	if response.Message != unmarshaled.Message {
		t.Errorf("Message mismatch: got %v, want %v", unmarshaled.Message, response.Message)
	}
	if response.AllowCopy != unmarshaled.AllowCopy {
		t.Errorf("AllowCopy mismatch: got %v, want %v", unmarshaled.AllowCopy, response.AllowCopy)
	}
}
