# JWT Structure

A JWT consists of three parts encoded in Base64URL format and separated by dots:

Header.Payload.Signature

For example:

eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZXhwIjoxNjgwMDAwMDAwfQ.8Gj_9bJjAqQ-5j3iCKMzVnlg-d1Kk-fXnOKC1Vt2fGc

The header identifies the algorithm used for signing:

// In Go, the header is typically handled by the JWT library
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

The payload contains claims about the user like ID, roles, and expiration time

// Creating claims in Go
claims := jwt.MapClaims{
    "sub": user.ID.String(),
    "username": user.Username,
    "exp": time.Now().Add(15 * time.Minute).Unix(),
}

The signature verifies the token hasn't been tampered with

// Signing the token with our secret
tokenString, err := token.SignedString([]byte(jwtSecret))

How Our JWT Flow Works
To understand how JWT fits into our Go authentication system, let's walk through the flow of a user logging in and accessing protected routes:

When a user successfully authenticates, our Go service:

Validates credentials against our database
Creates JWT with appropriate claims and expiration
Signs the token with a secret key
The client:

Stores the JWT (typically in localStorage or a secure cookie)
Includes the token in the Authorization header for subsequent requests
Authorization: Bearer eyJhbGciOiJIUzI1Ni...

Our middleware:

Extracts the JWT from the request header
Validates the signature using our secret key
Checks that the token hasn't expired
Extracts the user identity from claims
Adds the user ID to the request context
Since the token contains all necessary user information, our server can authenticate requests without maintaining session state or additional database queries.

The security of this system relies on keeping the signing key secret and using short-lived access tokens. If a token is compromised, it's only valid for a limited time, reducing the risk of unauthorized access.
