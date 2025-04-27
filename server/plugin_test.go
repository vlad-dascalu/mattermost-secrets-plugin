package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mattermost/mattermost-plugin-secrets/server/models"
	"github.com/mattermost/mattermost-plugin-secrets/server/store"
)

func setupTestPlugin(t *testing.T, mockSecretStore store.SecretStore) *Plugin {
	t.Helper()
	
	mockAPI := &plugintest.API{}
	
	mockAPI.On("LogError", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAPI.On("LogWarn", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAPI.On("LogInfo", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAPI.On("LogDebug", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	p := &Plugin{}
	p.SetAPI(mockAPI)
	p.secretStore = mockSecretStore
	
	p.setConfiguration(&configuration{
		SecretExpiryTime:     24,
		AllowCopyToClipboard: true,
	})
	
	return p
}

func TestPlugin_ExecuteCommand(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		commandArgs   *model.CommandArgs
		mockStore     func() store.SecretStore
		mockAPI       func(api *plugintest.API)
		expectedResp  *model.CommandResponse
		expectedError bool
	}{
		{
			name: "successfully creates a secret",
			commandArgs: &model.CommandArgs{
				Command:   "/secret This is a test secret",
				UserId:    "user1",
				ChannelId: "channel1",
				TeamId:    "team1",
				Username:  "username1",
			},
			mockStore: func() store.SecretStore {
				mockStore := &MockSecretStore{}
				mockStore.On("SaveSecret", mock.AnythingOfType("*models.Secret")).Return(nil)
				return mockStore
			},
			mockAPI: func(api *plugintest.API) {
				api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(&model.Post{}, nil)
			},
			expectedResp: &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Secret message created successfully.",
			},
			expectedError: false,
		},
		{
			name: "empty message",
			commandArgs: &model.CommandArgs{
				Command:   "/secret",
				UserId:    "user1",
				ChannelId: "channel1",
				TeamId:    "team1",
			},
			mockStore: func() store.SecretStore {
				return &MockSecretStore{}
			},
			mockAPI: func(api *plugintest.API) {},
			expectedResp: &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Please provide a message to be kept secret.",
			},
			expectedError: false,
		},
		{
			name: "error saving secret",
			commandArgs: &model.CommandArgs{
				Command:   "/secret This is a test secret",
				UserId:    "user1",
				ChannelId: "channel1",
				TeamId:    "team1",
				Username:  "username1",
			},
			mockStore: func() store.SecretStore {
				mockStore := &MockSecretStore{}
				mockStore.On("SaveSecret", mock.AnythingOfType("*models.Secret")).Return(errors.New("test error"))
				return mockStore
			},
			mockAPI: func(api *plugintest.API) {},
			expectedResp: &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Error creating secret: failed to save secret: test error",
			},
			expectedError: false,
		},
		{
			name: "error creating post",
			commandArgs: &model.CommandArgs{
				Command:   "/secret This is a test secret",
				UserId:    "user1",
				ChannelId: "channel1",
				TeamId:    "team1",
				Username:  "username1",
			},
			mockStore: func() store.SecretStore {
				mockStore := &MockSecretStore{}
				mockStore.On("SaveSecret", mock.AnythingOfType("*models.Secret")).Return(nil)
				return mockStore
			},
			mockAPI: func(api *plugintest.API) {
				api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, &model.AppError{Message: "test error"})
			},
			expectedResp: &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Error posting secret message: test error",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			mockAPI.On("LogError", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockAPI.On("LogWarn", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockAPI.On("LogInfo", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockAPI.On("LogDebug", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			
			// Apply any additional API mocks
			tt.mockAPI(mockAPI)
			
			p := &Plugin{}
			p.SetAPI(mockAPI)
			p.secretStore = tt.mockStore()
			p.botID = "bot1"
			
			p.setConfiguration(&configuration{
				SecretExpiryTime:     24,
				AllowCopyToClipboard: true,
			})
			
			resp, err := p.ExecuteCommand(&plugin.Context{}, tt.commandArgs)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}

func TestPlugin_HandleSecret(t *testing.T) {
	// Test cases
	tests := []struct {
		name               string
		method             string
		userID             string
		body               string
		mockStore          func() store.SecretStore
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:   "successfully creates a secret",
			method: http.MethodPost,
			userID: "user1",
			body:   `{"channel_id": "channel1", "message": "This is a test secret"}`,
			mockStore: func() store.SecretStore {
				mockStore := &MockSecretStore{}
				mockStore.On("SaveSecret", mock.AnythingOfType("*models.Secret")).Return(nil)
				return mockStore
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "missing channel_id",
			method: http.MethodPost,
			userID: "user1",
			body:   `{"message": "This is a test secret"}`,
			mockStore: func() store.SecretStore {
				return &MockSecretStore{}
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "channelId and message are required",
		},
		{
			name:   "missing message",
			method: http.MethodPost,
			userID: "user1",
			body:   `{"channel_id": "channel1"}`,
			mockStore: func() store.SecretStore {
				return &MockSecretStore{}
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "channelId and message are required",
		},
		{
			name:   "invalid JSON",
			method: http.MethodPost,
			userID: "user1",
			body:   `invalid JSON`,
			mockStore: func() store.SecretStore {
				return &MockSecretStore{}
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "error saving secret",
			method: http.MethodPost,
			userID: "user1",
			body:   `{"channel_id": "channel1", "message": "This is a test secret"}`,
			mockStore: func() store.SecretStore {
				mockStore := &MockSecretStore{}
				mockStore.On("SaveSecret", mock.AnythingOfType("*models.Secret")).Return(errors.New("test error"))
				return mockStore
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "method not allowed",
			method: http.MethodGet,
			userID: "user1",
			body:   `{}`,
			mockStore: func() store.SecretStore {
				return &MockSecretStore{}
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "unauthorized",
			method: http.MethodPost,
			userID: "",
			body:   `{"channel_id": "channel1", "message": "This is a test secret"}`,
			mockStore: func() store.SecretStore {
				return &MockSecretStore{}
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := setupTestPlugin(t, tt.mockStore())
			
			req, err := http.NewRequest(tt.method, "/api/v1/secrets", strings.NewReader(tt.body))
			assert.NoError(t, err)
			
			if tt.userID != "" {
				req.Header.Set("Mattermost-User-Id", tt.userID)
			}
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			p.handleSecret(w, req)
			
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestPlugin_MarkSecretAsViewed(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		secretID      string
		userID        string
		existingViews []string
		mockStore     func(secretID string, existingViews []string) store.SecretStore
		expectedError bool
	}{
		{
			name:          "successfully marks secret as viewed",
			secretID:      "secret1",
			userID:        "user1",
			existingViews: []string{},
			mockStore: func(secretID string, existingViews []string) store.SecretStore {
				mockStore := &MockSecretStore{}
				
				secret := &models.Secret{
					ID:       secretID,
					ViewedBy: existingViews,
				}
				
				// First GetSecret call
				mockStore.On("GetSecret", secretID).Return(secret, nil)
				
				// SaveSecret call should include user1 in ViewedBy
				expectedSecret := &models.Secret{
					ID:       secretID,
					ViewedBy: append(existingViews, "user1"),
				}
				mockStore.On("SaveSecret", mock.MatchedBy(func(s *models.Secret) bool {
					// Check if ViewedBy contains the expected users
					if len(s.ViewedBy) != len(expectedSecret.ViewedBy) {
						return false
					}
					
					for i, id := range s.ViewedBy {
						if id != expectedSecret.ViewedBy[i] {
							return false
						}
					}
					
					return s.ID == expectedSecret.ID
				})).Return(nil)
				
				return mockStore
			},
			expectedError: false,
		},
		{
			name:          "user already viewed secret",
			secretID:      "secret1",
			userID:        "user1",
			existingViews: []string{"user1"},
			mockStore: func(secretID string, existingViews []string) store.SecretStore {
				mockStore := &MockSecretStore{}
				
				secret := &models.Secret{
					ID:       secretID,
					ViewedBy: existingViews,
				}
				
				// GetSecret call
				mockStore.On("GetSecret", secretID).Return(secret, nil)
				
				return mockStore
			},
			expectedError: false,
		},
		{
			name:          "secret not found",
			secretID:      "secret1",
			userID:        "user1",
			existingViews: []string{},
			mockStore: func(secretID string, existingViews []string) store.SecretStore {
				mockStore := &MockSecretStore{}
				
				// GetSecret call - not found
				mockStore.On("GetSecret", secretID).Return(nil, nil)
				
				return mockStore
			},
			expectedError: true,
		},
		{
			name:          "error getting secret",
			secretID:      "secret1",
			userID:        "user1",
			existingViews: []string{},
			mockStore: func(secretID string, existingViews []string) store.SecretStore {
				mockStore := &MockSecretStore{}
				
				// GetSecret call - error
				mockStore.On("GetSecret", secretID).Return(nil, errors.New("test error"))
				
				return mockStore
			},
			expectedError: true,
		},
		{
			name:          "error saving secret",
			secretID:      "secret1",
			userID:        "user1",
			existingViews: []string{},
			mockStore: func(secretID string, existingViews []string) store.SecretStore {
				mockStore := &MockSecretStore{}
				
				secret := &models.Secret{
					ID:       secretID,
					ViewedBy: existingViews,
				}
				
				// GetSecret call
				mockStore.On("GetSecret", secretID).Return(secret, nil)
				
				// SaveSecret call - error
				mockStore.On("SaveSecret", mock.AnythingOfType("*models.Secret")).Return(errors.New("test error"))
				
				return mockStore
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := setupTestPlugin(t, tt.mockStore(tt.secretID, tt.existingViews))
			
			err := p.markSecretAsViewed(tt.secretID, tt.userID)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockSecretStore is a mock implementation of the SecretStore interface for testing
type MockSecretStore struct {
	mock.Mock
}

func (m *MockSecretStore) SaveSecret(secret *models.Secret) error {
	args := m.Called(secret)
	return args.Error(0)
}

func (m *MockSecretStore) GetSecret(id string) (*models.Secret, error) {
	args := m.Called(id)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*models.Secret), args.Error(1)
}

func (m *MockSecretStore) DeleteSecret(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSecretStore) ListExpiredSecrets() ([]*models.Secret, error) {
	args := m.Called()
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).([]*models.Secret), args.Error(1)
} 