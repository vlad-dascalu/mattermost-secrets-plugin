package store

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vlad-dascalu/mattermost-secrets-plugin/server/models"
)

func TestKVSecretStore_SaveSecret(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.Secret
		mockAPI   func(api *plugintest.API)
		expectErr bool
	}{
		{
			name: "successfully saves secret",
			secret: &models.Secret{
				ID:        "secret1",
				UserID:    "user1",
				ChannelID: "channel1",
				Message:   "test secret",
			},
			mockAPI: func(api *plugintest.API) {
				api.On("KVSet", mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "empty secret ID",
			secret: &models.Secret{
				ID: "",
			},
			mockAPI:   func(api *plugintest.API) {},
			expectErr: true,
		},
		{
			name: "error saving to KV store",
			secret: &models.Secret{
				ID: "secret1",
			},
			mockAPI: func(api *plugintest.API) {
				api.On("KVSet", mock.Anything, mock.Anything).Return(&model.AppError{Message: "error"})
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			tt.mockAPI(mockAPI)

			store := NewKVSecretStore(mockAPI)
			err := store.SaveSecret(tt.secret)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockAPI.AssertExpectations(t)
			}
		})
	}
}

func TestKVSecretStore_GetSecret(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockAPI   func(api *plugintest.API)
		expectErr bool
	}{
		{
			name: "successfully gets secret",
			id:   "secret1",
			mockAPI: func(api *plugintest.API) {
				secret := &models.Secret{
					ID:      "secret1",
					Message: "test secret",
				}
				data, _ := json.Marshal(secret)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(data, nil)
			},
			expectErr: false,
		},
		{
			name:      "empty secret ID",
			id:        "",
			mockAPI:   func(api *plugintest.API) {},
			expectErr: true,
		},
		{
			name: "secret not found",
			id:   "nonexistent",
			mockAPI: func(api *plugintest.API) {
				api.On("KVGet", SecretKeyPrefix+"nonexistent").Return(nil, nil)
			},
			expectErr: false,
		},
		{
			name: "error getting from KV store",
			id:   "secret1",
			mockAPI: func(api *plugintest.API) {
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(nil, &model.AppError{Message: "error"})
			},
			expectErr: true,
		},
		{
			name: "invalid JSON data",
			id:   "secret1",
			mockAPI: func(api *plugintest.API) {
				api.On("KVGet", SecretKeyPrefix+"secret1").Return([]byte("invalid json"), nil)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			tt.mockAPI(mockAPI)

			store := NewKVSecretStore(mockAPI)
			secret, err := store.GetSecret(tt.id)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, secret)
			} else {
				assert.NoError(t, err)
				mockAPI.AssertExpectations(t)
			}
		})
	}
}

func TestKVSecretStore_DeleteSecret(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockAPI   func(api *plugintest.API)
		expectErr bool
	}{
		{
			name: "successfully deletes secret",
			id:   "secret1",
			mockAPI: func(api *plugintest.API) {
				api.On("KVDelete", SecretKeyPrefix+"secret1").Return(nil)
			},
			expectErr: false,
		},
		{
			name:      "empty secret ID",
			id:        "",
			mockAPI:   func(api *plugintest.API) {},
			expectErr: true,
		},
		{
			name: "error deleting from KV store",
			id:   "secret1",
			mockAPI: func(api *plugintest.API) {
				api.On("KVDelete", SecretKeyPrefix+"secret1").Return(&model.AppError{Message: "error"})
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			tt.mockAPI(mockAPI)

			store := NewKVSecretStore(mockAPI)
			err := store.DeleteSecret(tt.id)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockAPI.AssertExpectations(t)
			}
		})
	}
}

func TestKVSecretStore_ListExpiredSecrets(t *testing.T) {
	tests := []struct {
		name      string
		mockAPI   func(api *plugintest.API)
		expectErr bool
	}{
		{
			name: "successfully lists expired secrets",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return([]string{SecretKeyPrefix + "secret1"}, nil)
				secret := &models.Secret{
					ID:        "secret1",
					ExpiresAt: 1, // Expired
				}
				data, _ := json.Marshal(secret)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(data, nil)
			},
			expectErr: false,
		},
		{
			name: "error listing from KV store",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return(nil, &model.AppError{Message: "error"})
			},
			expectErr: true,
		},
		{
			name: "error getting secret from KV store",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return([]string{SecretKeyPrefix + "secret1"}, nil)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(nil, &model.AppError{Message: "error"})
			},
			expectErr: false,
		},
		{
			name: "invalid JSON data",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return([]string{SecretKeyPrefix + "secret1"}, nil)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return([]byte("invalid json"), nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			// Set up LogError mocks for all possible error cases
			mockAPI.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			// Then set up other mocks
			tt.mockAPI(mockAPI)

			store := NewKVSecretStore(mockAPI)
			secrets, err := store.ListExpiredSecrets()

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, secrets)
			} else {
				assert.NoError(t, err)
				mockAPI.AssertExpectations(t)
			}
		})
	}
}

func TestKVSecretStore_GetAllSecrets(t *testing.T) {
	tests := []struct {
		name      string
		mockAPI   func(api *plugintest.API)
		expectErr bool
	}{
		{
			name: "successfully gets all secrets",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return([]string{SecretKeyPrefix + "secret1"}, nil)
				secret := &models.Secret{
					ID: "secret1",
				}
				data, _ := json.Marshal(secret)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(data, nil)
			},
			expectErr: false,
		},
		{
			name: "error listing from KV store",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return(nil, &model.AppError{Message: "error"})
			},
			expectErr: true,
		},
		{
			name: "error getting secret from KV store",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return([]string{SecretKeyPrefix + "secret1"}, nil)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return(nil, &model.AppError{Message: "error"})
			},
			expectErr: false,
		},
		{
			name: "invalid JSON data",
			mockAPI: func(api *plugintest.API) {
				api.On("KVList", 0, 1000).Return([]string{SecretKeyPrefix + "secret1"}, nil)
				api.On("KVGet", SecretKeyPrefix+"secret1").Return([]byte("invalid json"), nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			// Set up LogError mocks for all possible error cases
			mockAPI.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			// Then set up other mocks
			tt.mockAPI(mockAPI)

			store := NewKVSecretStore(mockAPI)
			secrets, err := store.GetAllSecrets()

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, secrets)
			} else {
				assert.NoError(t, err)
				mockAPI.AssertExpectations(t)
			}
		})
	}
}
