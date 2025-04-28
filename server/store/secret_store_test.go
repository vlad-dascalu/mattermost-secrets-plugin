package store

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mattermost/mattermost-plugin-secrets/server/models"
)

func TestKVSecretStore_SaveSecret(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		secret        *models.Secret
		mockAPI       func(api *plugintest.API, secret *models.Secret)
		expectedError bool
	}{
		{
			name: "successfully saves secret",
			secret: &models.Secret{
				ID:        "secret1",
				UserID:    "user1",
				ChannelID: "channel1",
				Message:   "This is a test secret",
				ViewedBy:  []string{},
				CreatedAt: 12345,
				ExpiresAt: 67890,
			},
			mockAPI: func(api *plugintest.API, secret *models.Secret) {
				expectedKey := SecretKeyPrefix + secret.ID
				serialized, _ := json.Marshal(secret)
				api.On("KVSet", expectedKey, serialized).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "error saving to KV store",
			secret: &models.Secret{
				ID:        "secret1",
				UserID:    "user1",
				ChannelID: "channel1",
				Message:   "This is a test secret",
				ViewedBy:  []string{},
				CreatedAt: 12345,
				ExpiresAt: 67890,
			},
			mockAPI: func(api *plugintest.API, secret *models.Secret) {
				expectedKey := SecretKeyPrefix + secret.ID
				serialized, _ := json.Marshal(secret)
				api.On("KVSet", expectedKey, serialized).Return(errors.New("test error"))
			},
			expectedError: true,
		},
		{
			name:   "empty secret ID",
			secret: &models.Secret{},
			mockAPI: func(api *plugintest.API, secret *models.Secret) {
				// No API calls expected
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}

			// Apply mocks
			tt.mockAPI(mockAPI, tt.secret)

			store := NewKVSecretStore(mockAPI)
			err := store.SaveSecret(tt.secret)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestKVSecretStore_GetSecret(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		secretID      string
		mockAPI       func(api *plugintest.API, secretID string)
		expectedError bool
		expectedNil   bool
	}{
		{
			name:     "successfully gets secret",
			secretID: "secret1",
			mockAPI: func(api *plugintest.API, secretID string) {
				secret := &models.Secret{
					ID:        secretID,
					UserID:    "user1",
					ChannelID: "channel1",
					Message:   "This is a test secret",
					ViewedBy:  []string{},
					CreatedAt: 12345,
					ExpiresAt: 67890,
				}

				expectedKey := SecretKeyPrefix + secretID
				serialized, _ := json.Marshal(secret)
				api.On("KVGet", expectedKey).Return(serialized, nil)
			},
			expectedError: false,
			expectedNil:   false,
		},
		{
			name:     "secret not found",
			secretID: "secret1",
			mockAPI: func(api *plugintest.API, secretID string) {
				expectedKey := SecretKeyPrefix + secretID
				api.On("KVGet", expectedKey).Return(nil, nil)
			},
			expectedError: false,
			expectedNil:   true,
		},
		{
			name:     "error getting from KV store",
			secretID: "secret1",
			mockAPI: func(api *plugintest.API, secretID string) {
				expectedKey := SecretKeyPrefix + secretID
				api.On("KVGet", expectedKey).Return(nil, errors.New("test error"))
			},
			expectedError: true,
			expectedNil:   true,
		},
		{
			name:     "invalid JSON",
			secretID: "secret1",
			mockAPI: func(api *plugintest.API, secretID string) {
				expectedKey := SecretKeyPrefix + secretID
				api.On("KVGet", expectedKey).Return([]byte("invalid JSON"), nil)
			},
			expectedError: true,
			expectedNil:   true,
		},
		{
			name:     "empty secret ID",
			secretID: "",
			mockAPI: func(api *plugintest.API, secretID string) {
				// No API calls expected
			},
			expectedError: true,
			expectedNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}

			// Apply mocks
			tt.mockAPI(mockAPI, tt.secretID)

			store := NewKVSecretStore(mockAPI)
			secret, err := store.GetSecret(tt.secretID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedNil {
				assert.Nil(t, secret)
			} else {
				assert.NotNil(t, secret)
				assert.Equal(t, tt.secretID, secret.ID)
			}

			// Verify mocks
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestKVSecretStore_DeleteSecret(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		secretID      string
		mockAPI       func(api *plugintest.API, secretID string)
		expectedError bool
	}{
		{
			name:     "successfully deletes secret",
			secretID: "secret1",
			mockAPI: func(api *plugintest.API, secretID string) {
				expectedKey := SecretKeyPrefix + secretID
				api.On("KVDelete", expectedKey).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "error deleting from KV store",
			secretID: "secret1",
			mockAPI: func(api *plugintest.API, secretID string) {
				expectedKey := SecretKeyPrefix + secretID
				api.On("KVDelete", expectedKey).Return(errors.New("test error"))
			},
			expectedError: true,
		},
		{
			name:     "empty secret ID",
			secretID: "",
			mockAPI: func(api *plugintest.API, secretID string) {
				// No API calls expected
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}

			// Apply mocks
			tt.mockAPI(mockAPI, tt.secretID)

			store := NewKVSecretStore(mockAPI)
			err := store.DeleteSecret(tt.secretID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestKVSecretStore_ListExpiredSecrets(t *testing.T) {
	// Current time in milliseconds
	now := int64(1000000)

	// Set up secrets
	secrets := []*models.Secret{
		{
			ID:        "secret1",
			ExpiresAt: now - 1000, // Expired 1 second ago
		},
		{
			ID:        "secret2",
			ExpiresAt: now + 1000, // Expires in 1 second
		},
		{
			ID:        "secret3",
			ExpiresAt: now - 2000, // Expired 2 seconds ago
		},
		{
			ID:        "secret4",
			ExpiresAt: 0, // Never expires
		},
	}

	// Set up keys
	keys := []string{
		SecretKeyPrefix + "secret1",
		SecretKeyPrefix + "secret2",
		SecretKeyPrefix + "secret3",
		SecretKeyPrefix + "secret4",
		"other_prefix_key",
	}

	// Test cases
	tests := []struct {
		name           string
		mockAPI        func(api *plugintest.API)
		expectedError  bool
		expectedCount  int
		expectedIDs    []string
		expectedMillis int64
	}{
		{
			name: "successfully lists expired secrets",
			mockAPI: func(api *plugintest.API) {
				// First, mock the KVList call
				api.On("KVList", 0, 1000).Return(keys, nil)

				// Then, mock each KVGet call for the secrets
				for _, secret := range secrets {
					key := SecretKeyPrefix + secret.ID
					serialized, _ := json.Marshal(secret)
					api.On("KVGet", key).Return(serialized, nil)
				}

				// Mock the current time
				models.GetMillis = func() int64 {
					return now
				}
			},
			expectedError: false,
			expectedCount: 2,                              // Expect 2 expired secrets
			expectedIDs:   []string{"secret1", "secret3"}, // The expired ones
		},
		{
			name: "error listing keys",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return(nil, errors.New("test error"))

				// Mock the current time
				models.GetMillis = func() int64 {
					return now
				}
			},
			expectedError: true,
		},
		{
			name: "error getting secret",
			mockAPI: func(api *plugintest.API) {
				// First, mock the KVList call
				api.On("KVList", 0, 1000).Return(keys, nil)

				// Mock an error for one of the secrets
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(nil, errors.New("test error"))

				// Mock successful gets for the other secrets
				for i := 1; i < len(secrets); i++ {
					secret := secrets[i]
					key := SecretKeyPrefix + secret.ID
					serialized, _ := json.Marshal(secret)
					api.On("KVGet", key).Return(serialized, nil)
				}

				// Mock the LogError call that will happen when a get fails
				api.On("LogError", mock.Anything, mock.Anything, mock.Anything).Return()

				// Mock the current time
				models.GetMillis = func() int64 {
					return now
				}
			},
			expectedError: false,
			expectedCount: 1, // Expect 1 expired secret (secret3)
			expectedIDs:   []string{"secret3"},
		},
		{
			name: "invalid JSON",
			mockAPI: func(api *plugintest.API) {
				// First, mock the KVList call
				api.On("KVList", 0, 1000).Return(keys, nil)

				// Mock invalid JSON for one of the secrets
				api.On("KVGet", SecretKeyPrefix+"secret1").Return([]byte("invalid JSON"), nil)

				// Mock successful gets for the other secrets
				for i := 1; i < len(secrets); i++ {
					secret := secrets[i]
					key := SecretKeyPrefix + secret.ID
					serialized, _ := json.Marshal(secret)
					api.On("KVGet", key).Return(serialized, nil)
				}

				// Mock the LogError call that will happen when a get fails
				api.On("LogError", mock.Anything, mock.Anything, mock.Anything).Return()

				// Mock the current time
				models.GetMillis = func() int64 {
					return now
				}
			},
			expectedError: false,
			expectedCount: 1, // Expect 1 expired secret (secret3)
			expectedIDs:   []string{"secret3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}

			// Apply mocks
			tt.mockAPI(mockAPI)

			// Reset GetMillis after the test
			defer func() {
				models.GetMillis = func() int64 {
					return 0
				}
			}()

			store := NewKVSecretStore(mockAPI)
			secrets, err := store.ListExpiredSecrets()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(secrets))

				// Check that each expected ID is in the results
				for _, expectedID := range tt.expectedIDs {
					found := false
					for _, secret := range secrets {
						if secret.ID == expectedID {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected to find secret with ID %s", expectedID)
				}
			}

			// Verify mocks
			mockAPI.AssertExpectations(t)
		})
	}
}
