package payment_intent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"example.com/sample-repo/qr_pay/jwt_generation"
)

// ClientConfig holds configuration for the PayNet DuitNow Pay payment intent client.
type ClientConfig struct {
	// BaseURL is the PayNet DuitNow Pay API base URL (e.g. https://certification.api.developer.inet.paynet.my). Path /v1/payment/intent is appended.
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

// SampleRequest returns a sample PaymentIntentRequest for one-time payment (dataType "01") per PayNet docs.
func SampleRequest() PaymentIntentRequest {
	return PaymentIntentRequest{
		DataType:        "01",
		TransactionFlow: "01",
		CheckoutID:      "a7e2ed2a-b088-4495-8cf4-88da08f644f2",
		SourceOfFunds:   []string{"01"},
		Amount:          "10.00",
		MerchantName:    "Shop Name Sdn Bhd.",
		MerchantRefID:   "M0000569",
		Merchant:        Merchant{ProductID: "C00000569"},
		Customer: Customer{
			Name:               "Walter Mitty",
			IdentityValidation: "00",
			IdentificationType: "01",
			Identification:     "840312145594",
		},
		Language: "en",
	}
}

// CreatePaymentIntent sends a POST request to PayNet's /v1/payment/intent and returns the response.
// Authorization uses JWT with the request body as the "data" claim per DuitNow Pay API Authentication.
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/payment-intent#send-the-payment-intent-request
func CreatePaymentIntent(cfg ClientConfig, req PaymentIntentRequest) (*PaymentIntentResponse, int, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

	// JWT "data" claim must be the request payload. Use the same JSON as body.
	var dataClaim interface{}
	if err := json.Unmarshal(bodyBytes, &dataClaim); err != nil {
		return nil, 0, fmt.Errorf("unmarshal for data claim: %w", err)
	}

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
	url := baseURL + "/v1/payment/intent"

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	log.Printf("[payment_intent] --- Outgoing request to PayNet ---")
	log.Printf("[payment_intent] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[payment_intent] Body:\n%s", string(bodyIndent))
	log.Printf("[payment_intent] -----------------------------------------")

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

	log.Printf("[payment_intent] --- Response from PayNet ---")
	log.Printf("[payment_intent] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[payment_intent] Body:\n%s", string(respBody))
	log.Printf("[payment_intent] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var intentResp PaymentIntentResponse
	if err := json.Unmarshal(respBody, &intentResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &intentResp, resp.StatusCode, nil
}
