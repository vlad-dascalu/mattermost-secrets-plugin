package store

import (
	"encoding/json"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"

	"github.com/vlad-dascalu/mattermost-secrets-plugin/server/models"
)

const (
	// SecretKeyPrefix is the KV store prefix for secret objects
	SecretKeyPrefix = "secret_"
)

// SecretStore defines the interface for storing and retrieving secrets
type SecretStore interface {
	// SaveSecret stores a secret in the KV store
	SaveSecret(secret *models.Secret) error

	// GetSecret retrieves a secret from the KV store by ID
	GetSecret(id string) (*models.Secret, error)

	// DeleteSecret removes a secret from the KV store
	DeleteSecret(id string) error

	// ListExpiredSecrets returns a list of secrets that have expired
	ListExpiredSecrets() ([]*models.Secret, error)

	// GetAllSecrets returns all secrets in the store
	GetAllSecrets() ([]*models.Secret, error)
}

// KVSecretStore implements the SecretStore interface using the plugin KV store
type KVSecretStore struct {
	api plugin.API
}

// NewKVSecretStore creates a new KVSecretStore
func NewKVSecretStore(api plugin.API) *KVSecretStore {
	return &KVSecretStore{
		api: api,
	}
}

// SaveSecret stores a secret in the KV store
func (s *KVSecretStore) SaveSecret(secret *models.Secret) error {
	if secret.ID == "" {
		return errors.New("secret ID cannot be empty")
	}

	data, err := json.Marshal(secret)
	if err != nil {
		return errors.Wrap(err, "failed to marshal secret")
	}

	key := SecretKeyPrefix + secret.ID

	if err := s.api.KVSet(key, data); err != nil {
		return errors.Wrap(err, "failed to store secret in KV store")
	}

	return nil
}

// GetSecret retrieves a secret from the KV store by ID
func (s *KVSecretStore) GetSecret(id string) (*models.Secret, error) {
	if id == "" {
		return nil, errors.New("secret ID cannot be empty")
	}

	key := SecretKeyPrefix + id

	data, appErr := s.api.KVGet(key)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get secret from KV store")
	}

	if data == nil {
		return nil, nil
	}

	var secret models.Secret
	if err := json.Unmarshal(data, &secret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal secret")
	}

	return &secret, nil
}

// DeleteSecret removes a secret from the KV store
func (s *KVSecretStore) DeleteSecret(id string) error {
	if id == "" {
		return errors.New("secret ID cannot be empty")
	}

	key := SecretKeyPrefix + id

	if appErr := s.api.KVDelete(key); appErr != nil {
		return errors.Wrap(appErr, "failed to delete secret from KV store")
	}

	return nil
}

// ListExpiredSecrets returns a list of secrets that have expired
// This is a simple implementation that requires scanning all secrets
func (s *KVSecretStore) ListExpiredSecrets() ([]*models.Secret, error) {
	var expired []*models.Secret

	// Get all keys with our prefix
	keys, appErr := s.api.KVList(0, 1000)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to list secrets from KV store")
	}

	for _, key := range keys {
		// Skip keys that don't match our prefix
		if len(key) <= len(SecretKeyPrefix) || key[:len(SecretKeyPrefix)] != SecretKeyPrefix {
			continue
		}

		// Get the secret
		data, appErr := s.api.KVGet(key)
		if appErr != nil {
			s.api.LogError("Failed to get secret", "key", key, "error", appErr.Error())
			continue
		}

		if data == nil {
			continue
		}

		var secret models.Secret
		if err := json.Unmarshal(data, &secret); err != nil {
			s.api.LogError("Failed to unmarshal secret", "key", key, "error", err.Error())
			continue
		}

		// Check if the secret has expired
		if secret.ExpiresAt > 0 && secret.ExpiresAt < models.GetMillis() {
			expired = append(expired, &secret)
		}
	}

	return expired, nil
}

// GetAllSecrets returns all secrets in the KV store
func (s *KVSecretStore) GetAllSecrets() ([]*models.Secret, error) {
	var secrets []*models.Secret

	// Get all keys with our prefix
	keys, appErr := s.api.KVList(0, 1000)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to list secrets from KV store")
	}

	for _, key := range keys {
		// Skip keys that don't match our prefix
		if len(key) <= len(SecretKeyPrefix) || key[:len(SecretKeyPrefix)] != SecretKeyPrefix {
			continue
		}

		// Get the secret
		data, appErr := s.api.KVGet(key)
		if appErr != nil {
			s.api.LogError("Failed to get secret", "key", key, "error", appErr.Error())
			continue
		}

		if data == nil {
			continue
		}

		var secret models.Secret
		if err := json.Unmarshal(data, &secret); err != nil {
			s.api.LogError("Failed to unmarshal secret", "key", key, "error", err.Error())
			continue
		}

		secrets = append(secrets, &secret)
	}

	return secrets, nil
}
