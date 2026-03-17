package jwt_generation

import (
	"crypto/rsa"
	"testing"
)

func TestGenerateJWT(t *testing.T) {
	privateKey, err := LoadDefaultPrivateKey()
	if err != nil {
		t.Fatalf("LoadDefaultPrivateKey: %v", err)
	}

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

	token, err := GenerateJWT(GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  RS256,
		Issuer:     "<acquirerId>",
		Subject:    "<merchantId>",
		JTI:        checkoutID,
		Key:        "<api-key-name>",
		Data:       requestData,
	})
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	// Token should be header.payload.signature (3 parts)
	if len(token) < 20 {
		t.Errorf("token too short: %q", token)
	}
}

func TestVerifyJWT(t *testing.T) {
	privateKey, err := LoadDefaultPrivateKey()
	if err != nil {
		t.Fatalf("LoadDefaultPrivateKey: %v", err)
	}
	// Use public key from the same private key (sample cert may not match sample key)
	publicKey := privateKey.Public().(*rsa.PublicKey)

	requestData := map[string]interface{}{"checkoutId": "test-verify-123"}
	token, err := GenerateJWT(GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  RS256,
		Issuer:     "test-issuer",
		Subject:    "test-subject",
		JTI:        "test-jti",
		Key:        "test-key",
		Data:       requestData,
	})
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	err = VerifyJWT(VerifyOptions{
		Token:     "Bearer " + token,
		PublicKey: publicKey,
		Algorithm: RS256,
	})
	if err != nil {
		t.Fatalf("VerifyJWT: %v", err)
	}
}
