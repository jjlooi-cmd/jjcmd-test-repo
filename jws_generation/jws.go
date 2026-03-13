// Package jws_generation implements PayNet DuitNow JWS (JSON Web Signature) for API authentication.
// Ref: https://docs.developer.paynet.my/docs/duitnow-transfer/integration/security-&-encryption/message-signature/JSON-web-signature
package jws_generation

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Algorithm is the JWS signing algorithm (RS256 or RS512).
type Algorithm string

const (
	RS256 Algorithm = "RS256"
	RS512 Algorithm = "RS512"
)

func (a Algorithm) hash() crypto.Hash {
	switch a {
	case RS512:
		return crypto.SHA512
	default:
		return crypto.SHA256
	}
}

// GenerateOptions holds parameters for JWS generation.
type GenerateOptions struct {
	// PrivateKey is the RSA private key used to sign the JWS.
	PrivateKey *rsa.PrivateKey
	// Algorithm is RS256 or RS512 (default RS256).
	Algorithm Algorithm
	// Issuer is the issuer claim (e.g. client BIC code).
	Issuer string
	// BusinessMessageID is used as jti (JWT ID).
	BusinessMessageID string
	// CredentialKey is the JWT Credential Key assigned during onboarding.
	CredentialKey string
	// PayloadForHash is the exact request body used to compute "ds".
	// Must be minified JSON (no extra spaces). Use SHA-256 of empty string if body is empty.
	PayloadForHash []byte
	// ValidDuration is how long the token is valid (default 15 minutes).
	ValidDuration time.Duration
}

// GenerateJWS builds a JWS token for a PayNet API request.
// Place the result in the Authorization header as: "Bearer <token>".
func GenerateJWS(opts GenerateOptions) (string, error) {
	if opts.PrivateKey == nil {
		return "", fmt.Errorf("jws: private key is required")
	}
	if opts.Algorithm == "" {
		opts.Algorithm = RS256
	}
	if opts.ValidDuration == 0 {
		opts.ValidDuration = 15 * time.Minute
	}

	now := time.Now().Unix()
	exp := time.Now().Add(opts.ValidDuration).Unix()

	// 1. Compute ds = SHA-256 of (minified) request payload
	dsInput := opts.PayloadForHash
	if dsInput == nil {
		dsInput = []byte("")
	}
	h := sha256.Sum256(dsInput)
	ds := fmt.Sprintf("%x", h)

	// 2. JWS header: alg, typ
	header := map[string]string{
		"alg": string(opts.Algorithm),
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("jws: marshal header: %w", err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)


	if opts.CredentialKey == "" {
		opts.CredentialKey = "64feb830"
	}
	// 3. JWS body (claims): iss, iat, exp, key, jti, ds
	body := map[string]interface{}{
		"iss": opts.Issuer,
		"iat": now,
		"exp": exp,
		"key": opts.CredentialKey,
		"jti": opts.BusinessMessageID,
		"ds":  ds,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("jws: marshal body: %w", err)
	}
	bodyB64 := base64.RawURLEncoding.EncodeToString(bodyJSON)

	// 4. Sign signingInput = header + "." + body
	signingInput := headerB64 + "." + bodyB64
	hashFn := opts.Algorithm.hash()
	hasher := hashFn.New()
	hasher.Write([]byte(signingInput))
	digest := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, opts.PrivateKey, hashFn, digest)
	if err != nil {
		return "", fmt.Errorf("jws: sign: %w", err)
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(signature)

	// 5. Full token
	return signingInput + "." + sigB64, nil
}

// VerifyOptions holds parameters for JWS verification (e.g. API response).
type VerifyOptions struct {
	// Token is the JWS token (with or without "Bearer " prefix).
	Token string
	// PublicKey is the RSA public key (e.g. from PayNet certificate) to verify the signature.
	PublicKey *rsa.PublicKey
	// Algorithm used in the token (RS256 or RS512).
	Algorithm Algorithm
	// PayloadForHash is the actual API response body (minified) to compare with "ds" in the token.
	PayloadForHash []byte
}

// VerifyJWS verifies a JWS token from a PayNet API response.
// It checks the signature and that the payload hash matches the "ds" claim.
func VerifyJWS(opts VerifyOptions) error {
	token := strings.TrimSpace(opts.Token)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("jws: invalid token, expected 3 parts, got %d", len(parts))
	}

	headerB64, bodyB64, sigB64 := parts[0], parts[1], parts[2]
	signingInput := headerB64 + "." + bodyB64

	// Decode and verify signature
	sigBytes, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return fmt.Errorf("jws: decode signature: %w", err)
	}
	hashFn := opts.Algorithm.hash()
	hasher := hashFn.New()
	hasher.Write([]byte(signingInput))
	digest := hasher.Sum(nil)
	if err := rsa.VerifyPKCS1v15(opts.PublicKey, hashFn, digest, sigBytes); err != nil {
		return fmt.Errorf("jws: signature verification failed: %w", err)
	}

	// Decode body and check exp + ds
	bodyJSON, err := base64.RawURLEncoding.DecodeString(bodyB64)
	if err != nil {
		return fmt.Errorf("jws: decode body: %w", err)
	}
	var claims struct {
		Exp int64  `json:"exp"`
		DS  string `json:"ds"`
	}
	if err := json.Unmarshal(bodyJSON, &claims); err != nil {
		return fmt.Errorf("jws: parse body: %w", err)
	}
	if time.Now().Unix() > claims.Exp {
		return fmt.Errorf("jws: token expired (exp=%d)", claims.Exp)
	}

	// Compare ds with SHA-256 of actual payload
	payload := opts.PayloadForHash
	if payload == nil {
		payload = []byte("")
	}
	h := sha256.Sum256(payload)
	expectedDS := fmt.Sprintf("%x", h)
	if claims.DS != expectedDS {
		return fmt.Errorf("jws: payload hash mismatch (ds)")
	}
	return nil
}
