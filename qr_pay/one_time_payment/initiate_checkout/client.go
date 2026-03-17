package initiate_checkout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"example.com/sample-repo/qr_pay/common_header"
	"example.com/sample-repo/qr_pay/jwt_generation"
)

// API Documentation:
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-checkout

// ClientConfig holds configuration for the PayNet DuitNow Pay initiate checkout client.
type ClientConfig struct {
	// BaseURL is the PayNet DuitNow Pay API base URL (e.g. https://certification.api.developer.inet.paynet.my). Path /v1/bw/checkout is appended.
	BaseURL string
	// JWT Issuer (iss): BIC code assigned during onboarding.
	JWTIssuer string
	// JWT Subject (sub): Merchant ID from merchant registration.
	JWTSubject string
	// JWT Key (key): Project ID assigned during onboarding.
	JWTKey string
}

// DefaultClientConfig returns a config with placeholder JWT claims. Override BaseURL and JWT fields for your environment.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		BaseURL:    "https://certification.api.developer.inet.paynet.my/pay-guard",
		JWTIssuer:  "MBBEMYKL",
		JWTSubject: "M0000569",
		JWTKey:     "A46548895",
	}
}

// SampleRequest returns a sample InitiateCheckoutRequest per PayNet docs (self-hosted page initiate checkout).
func SampleRequest() InitiateCheckoutRequest {
	return InitiateCheckoutRequest{
		CheckoutID:      "a7e2ed2a-b088-4495-8cf4-88da08f644f2",
		TransactionFlow: "01",
		Issuer:          "Affin Bank",
		SourceOfFunds:   []string{"01"},
		Amount:          "10.00",
		Merchant:        Merchant{ProductID: "P00000201"},
		MerchantName:    "Shop Name Sdn Bhd.",
		MerchantRefID:   "ref12345678",
		Customer: Customer{
			Name:               "Walter Mitty",
			IdentificationType: "05",
			Identification:     "+60123456788",
			IdentityValidation: "00",
		},
	}
}

// InitiateCheckout sends a POST request to PayNet's /v1/bw/checkout and returns the response.
// Authorization uses JWT with the request body as the "data" claim per DuitNow Pay API Authentication.
// The endToEndIdSignature in the response is used to construct the browser redirection to the bank.
func InitiateCheckout(cfg ClientConfig, req InitiateCheckoutRequest) (*InitiateCheckoutResponse, int, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

	// JWT "data" claim must match the request body exactly (PayNet may compare).
	dataClaim := json.RawMessage(bodyBytes)

	jti := strings.TrimSpace(req.CheckoutID)
	if jti == "" {
		jti = "550e8400-e29b-41d4-a716-446655440000"
	}

	token, err := jwt_generation.GenerateJWT(jwt_generation.GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  jwt_generation.RS256,
		Issuer:     cfg.JWTIssuer,
		Subject:    cfg.JWTSubject,
		JTI:        jti,
		Key:        cfg.JWTKey,
		Data:       dataClaim,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	url := baseURL + "/v1/bw/checkout"

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[initiate_checkout] --- Outgoing request to PayNet ---")
	log.Printf("[initiate_checkout] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[initiate_checkout] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[initiate_checkout]   %s: %s", k, strings.Join(v, ", "))
	}
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[initiate_checkout] Body:\n%s", string(bodyIndent))
	log.Printf("[initiate_checkout] -----------------------------------------")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}

	log.Printf("[initiate_checkout] --- Response from PayNet ---")
	log.Printf("[initiate_checkout] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[initiate_checkout] Body:\n%s", string(respBody))
	log.Printf("[initiate_checkout] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var checkoutResp InitiateCheckoutResponse
	if err := json.Unmarshal(respBody, &checkoutResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &checkoutResp, resp.StatusCode, nil
}
