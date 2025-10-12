package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Hashpassword creates a bcrypt hash from the given plaintext password.
func Hashpassword(password string) (string, error) {
	// the cost determines the computational complexity of the hashing process
	// higher cost means more security but also more processing time
	// default cost is 10, we can adjust based on our security needs and performance considerations
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword checks if the provided plaintext password matches the stored bcrypt hash.
func VerifyPassword(hashedPassword, providedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}
