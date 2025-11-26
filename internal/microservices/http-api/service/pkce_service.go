package service

// PKCE = Proof Key for Code Exchange
// PKCEService defines methods for handling PKCE-related operations
// such as generating and validating code challenges and verifiers.

type PKCEService interface {
	// gen a 32-byte random value and encode with URL-safe base64
	GenerateCodeVerifier() (string, error)
	// compute SHA256(verifier) and base64url-encode the result.
	GenerateCodeChallenge(verifier string) string
	// Verify that the provided verifier matches the stored challenge
	VerifyCodeChallenge(verifier, challenge string) bool
}
