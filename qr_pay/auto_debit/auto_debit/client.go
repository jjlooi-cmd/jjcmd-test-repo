package auto_debit

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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/duitnow-autodebit

// ClientConfig holds configuration for the PayNet DuitNow Pay AutoDebit client.
type ClientConfig struct {
	BaseURL    string
	JWTIssuer  string
	JWTSubject string
	JWTKey     string
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

// SampleRequest returns a sample AutoDebitRequest per PayNet docs.
func SampleRequest() AutoDebitRequest {
	return AutoDebitRequest{
		CheckoutID:          "a7e2ed2a-b088-4495-8cf4-88da08f644f2",
		ConsentID:           "M00002010012700006",
		Amount:              "10.00",
		MerchantReferenceID: "REF0001234556",
	}
}

// InitiateAutoDebit sends a POST request to PayNet's /v1/bw/autodebit to trigger an AutoDebit payment using the given consent.
// Authorization uses JWT with the request body as the "data" claim per DuitNow Pay API Authentication.
func InitiateAutoDebit(cfg ClientConfig, req AutoDebitRequest) (*AutoDebitResponse, int, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

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
	url := baseURL + "/v1/bw/autodebit"

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[auto_debit] --- Outgoing request to PayNet ---")
	log.Printf("[auto_debit] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[auto_debit] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[auto_debit]   %s: %s", k, strings.Join(v, ", "))
	}
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[auto_debit] Body:\n%s", string(bodyIndent))
	log.Printf("[auto_debit] -----------------------------------------")

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

	log.Printf("[auto_debit] --- Response from PayNet ---")
	log.Printf("[auto_debit] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[auto_debit] Body:\n%s", string(respBody))
	log.Printf("[auto_debit] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var autodebitResp AutoDebitResponse
	if err := json.Unmarshal(respBody, &autodebitResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &autodebitResp, resp.StatusCode, nil
}
