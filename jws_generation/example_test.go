package jws_generation_test

import (
	"encoding/json"
	"fmt"
	"log"

	"example.com/sample-repo/jws_generation"
)

// Example_generateJWS shows how to generate a JWS token for a PayNet API request.
func Example_generateJWS() {
	// Load private key from sample_private_key.key in the same folder (or use LoadPrivateKey(path) for a custom path).
	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Request body that will be sent in the API call (minified for consistent hash).
	requestBody := map[string]interface{}{
		"data": map[string]interface{}{
			"businessMessageId": "20230412BOEEMYK1000ORB00000001",
			"clientMessage":     "Client hello",
		},
	}
	payloadBytes, _ := json.Marshal(requestBody) // minified, no extra spaces

	token, err := jws_generation.GenerateJWS(jws_generation.GenerateOptions{
		PrivateKey:         privateKey,
		Algorithm:          jws_generation.RS256,
		Issuer:             "BOEEMYK1",
		BusinessMessageID:  "20230412BOEEMYK1000ORB00000001",
		CredentialKey:      "ERTqafGRyt35MAKX5pBMU",
		PayloadForHash:     payloadBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Use in HTTP request: Authorization: Bearer <token>
	_ = token
	fmt.Printf("Authorization: Bearer %s...\n", token[:50])
}

// Example_verifyJWS shows how to verify a JWS token from a PayNet API response.
func Example_verifyJWS() {
	// Load certificate from sample_external_certificate.cer in the same folder (or use LoadCertificate(path) for a custom path).
	publicKey, _, err := jws_generation.LoadDefaultCertificate()
	if err != nil {
		log.Fatal(err)
	}

	tokenFromResponse := "Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9..."
	responseBody := []byte(`{"data":{"businessMessageId":"20230412BOEEMYK1000ORB00000001"}}`)

	err = jws_generation.VerifyJWS(jws_generation.VerifyOptions{
		Token:          tokenFromResponse,
		PublicKey:      publicKey,
		Algorithm:      jws_generation.RS256,
		PayloadForHash: responseBody,
	})
	if err != nil {
		log.Fatal("JWS verification failed:", err)
	}
	fmt.Println("JWS valid")
}
