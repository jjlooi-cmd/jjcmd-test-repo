package jwt_generation_test

import (
	"crypto/rsa"
	"fmt"
	"log"

	"example.com/sample-repo/qr_pay/jwt_generation"
)

// Example_generateJWT shows how to generate a JWT token for a PayNet DuitNow Pay API request.
func Example_generateJWT() {
	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Request body (data claim): same payload you send in the API request body.
	// Example matches PayNet Pay API Authentication sample (checkoutId, amount, merchant, etc.).
	checkoutID := "550e8400-e29b-41d4-a716-446655440000"
	requestData := map[string]interface{}{
		"checkoutId": checkoutID,
		"amount":     "10.00",
		"merchant":   map[string]interface{}{"productId": "P00000205"},
		"merchantReferenceId": "ref20240124T073716",
		"customer": map[string]interface{}{
			"name":                 "Walter Mitty",
			"identityValidation":   "00",
			"identificationType":  "05",
			"identification":      "+60123456789",
		},
		"consent": map[string]interface{}{
			"maxAmount":     "100.00",
			"effectiveDate": "2024-01-24",
			"expiryDate":    "2024-04-24",
			"frequency":     "01",
		},
	}

	token, err := jwt_generation.GenerateJWT(jwt_generation.GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  jwt_generation.RS256,
		Issuer:     "<acquirerId>",   // BIC code from onboarding
		Subject:    "<merchantId>",    // Merchant ID from registration
		JTI:        checkoutID,       // UUID v4
		Key:        "<api-key-name>", // Project ID from onboarding
		Data:       requestData,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Use in HTTP request: Authorization: Bearer <token>
	fmt.Printf("Authorization: Bearer %s...\n", token[:min(50, len(token))])
}

// Example_verifyJWT shows how to verify a JWT from a PayNet DuitNow Pay API response.
// This example generates a token then verifies it; in production you'd verify a token from the API response.
func Example_verifyJWT() {
	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public() // or use LoadDefaultCertificate() for PayNet's cert

	token, err := jwt_generation.GenerateJWT(jwt_generation.GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  jwt_generation.RS256,
		Issuer:     "example-issuer",
		Subject:    "example-subject",
		JTI:        "example-jti",
		Key:        "example-key",
		Data:       map[string]interface{}{"checkoutId": "example"},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = jwt_generation.VerifyJWT(jwt_generation.VerifyOptions{
		Token:     "Bearer " + token,
		PublicKey: publicKey.(*rsa.PublicKey),
		Algorithm: jwt_generation.RS256,
	})
	if err != nil {
		log.Fatal("JWT verification failed:", err)
	}
	fmt.Println("JWT valid")
}
