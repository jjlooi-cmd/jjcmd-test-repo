package enquire_refund

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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/refund-status

// ClientConfig holds configuration for the PayNet DuitNow Pay Enquire Refund Status client.
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

// EnquireRefundStatus sends a GET request to PayNet's /v1/bw/refund?refundId=... to retrieve the status of a refund.
// Should only be performed once and at least one hour after the initial refund request (when webhook update refund status fails).
// Authorization uses JWT; for GET with no body, the "data" claim is omitted per DuitNow Pay API Authentication.
func EnquireRefundStatus(cfg ClientConfig, refundId string) (*EnquireRefundStatusResponse, int, error) {
	if refundId == "" {
		return nil, 0, fmt.Errorf("refundId is required")
	}

	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

	jti := strings.TrimSpace(refundId)
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
		Data:       nil, // no body for GET; do not include "data" claim
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	rawURL := baseURL + "/v1/bw/refund"
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, 0, fmt.Errorf("parse URL: %w", err)
	}
	q := u.Query()
	q.Set("refundId", refundId)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[enquire_refund] --- Outgoing request to PayNet ---")
	log.Printf("[enquire_refund] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[enquire_refund] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[enquire_refund]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[enquire_refund] -----------------------------------------")

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

	log.Printf("[enquire_refund] --- Response from PayNet ---")
	log.Printf("[enquire_refund] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[enquire_refund] Body:\n%s", string(respBody))
	log.Printf("[enquire_refund] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResp EnquireRefundStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &statusResp, resp.StatusCode, nil
}

// SampleRefundId returns a sample refundId for trigger/testing (as in the API spec example).
func SampleRefundId() string {
	return "f36b4c31-44b2-40f2-820a-7fbd081cae9b"
}
