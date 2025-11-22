package authentication

// KeyString holds the key strings used for authentication headers and tokens, on the client side.
import (
	"encoding/json"
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "mangahub-cli"
	tokenKey    = "auth_tokens"
)

// ErrNotLoggedIn is returned when credentials are not found
var ErrNotLoggedIn = errors.New("not logged in")

type StoredCredentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
	ExpiresAt    int64  `json:"expires_at"`
}

// StoreTokens stores authentication tokens in the system keyring
func StoreTokens(creds *StoredCredentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	err = keyring.Set(serviceName, tokenKey, string(data))
	if err != nil {
		return errors.New("failed to store credentials in system keyring: " + err.Error())
	}
	return nil
}

// GetTokens retrieves authentication tokens from the system keyring
func GetTokens() (*StoredCredentials, error) {
	value, err := keyring.Get(serviceName, tokenKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrNotLoggedIn
		}
		return nil, errors.New("failed to retrieve credentials from system keyring: " + err.Error())
	}

	var creds StoredCredentials
	if err := json.Unmarshal([]byte(value), &creds); err != nil {
		return nil, errors.New("corrupted credentials data, please login again")
	}
	return &creds, nil
}

// DeleteTokens removes authentication tokens from the system keyring
func DeleteTokens() error {
	err := keyring.Delete(serviceName, tokenKey)
	if err != nil && err != keyring.ErrNotFound {
		return errors.New("failed to delete credentials from system keyring: " + err.Error())
	}
	return nil
}
