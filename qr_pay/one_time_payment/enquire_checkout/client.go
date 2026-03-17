package enquire_checkout

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"example.com/sample-repo/qr_pay/common_header"
	"example.com/sample-repo/qr_pay/jwt_generation"
)

// API Documentation:
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/checkout-details

// ClientConfig holds configuration for the PayNet DuitNow Pay Enquire Checkout Details client.
type ClientConfig struct {
	// BaseURL is the PayNet DuitNow Pay API base URL (e.g. https://certification.api.developer.inet.paynet.my/pay-guard). Path /v1/bw/checkout is appended.
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

// EnquireCheckout sends a GET request to PayNet's /v1/bw/checkout?endToEndId=... to retrieve checkout details.
// endToEndId is the end-to-end ID from the response of Initiate Payment Intent or Initiate Checkout (used when webhook failed).
// Authorization uses JWT; for GET with no body, the "data" claim is omitted per DuitNow Pay API Authentication.
func EnquireCheckout(cfg ClientConfig, endToEndId string) (*EnquireCheckoutResponse, int, error) {
	if endToEndId == "" {
		return nil, 0, fmt.Errorf("endToEndId is required")
	}

	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

	token, err := jwt_generation.GenerateJWT(jwt_generation.GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  jwt_generation.RS256,
		Issuer:     cfg.JWTIssuer,
		Subject:    cfg.JWTSubject,
		JTI:        "550e8400-e29b-41d4-a716-446655440001",
		Key:        cfg.JWTKey,
		Data:       nil, // no body for GET; do not include "data" claim
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	rawURL := baseURL + "/v1/bw/checkout"
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, 0, fmt.Errorf("parse URL: %w", err)
	}
	q := u.Query()
	q.Set("endToEndId", endToEndId)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[enquire_checkout] --- Outgoing request to PayNet ---")
	log.Printf("[enquire_checkout] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[enquire_checkout] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[enquire_checkout]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[enquire_checkout] -----------------------------------------")

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

	log.Printf("[enquire_checkout] --- Response from PayNet ---")
	log.Printf("[enquire_checkout] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[enquire_checkout] Body:\n%s", string(respBody))
	log.Printf("[enquire_checkout] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var checkoutResp EnquireCheckoutResponse
	if err := json.Unmarshal(respBody, &checkoutResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &checkoutResp, resp.StatusCode, nil
}

// SampleEndToEndId returns a sample endToEndId for trigger/testing (as in the API spec example).
func SampleEndToEndId() string {
	return "20240724M0037091861OBW05004745"
}
