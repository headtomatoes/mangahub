package authentication

// KeyString holds the key strings used for authentication headers and tokens, on the client side.
import (
	"encoding/json"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "mangahub-cli"
	tokenKey    = "auth_tokens"
)

type StoredCredentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
	ExpiresAt    int64  `json:"expires_at"`
}

func StoreTokens(creds *StoredCredentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return keyring.Set(serviceName, tokenKey, string(data))
}

func GetTokens() (*StoredCredentials, error) {
	value, err := keyring.Get(serviceName, tokenKey)
	if err != nil {
		return nil, err
	}

	var creds StoredCredentials
	if err := json.Unmarshal([]byte(value), &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

func DeleteTokens() error {
	return keyring.Delete(serviceName, tokenKey)
}
