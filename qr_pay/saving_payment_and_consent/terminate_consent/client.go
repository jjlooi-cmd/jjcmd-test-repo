package terminate_consent

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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/terminate-consent

// ClientConfig holds configuration for the PayNet DuitNow Pay Terminate Consent client.
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

// TerminateConsent sends a DELETE request to PayNet's /v1/bw/consent?consentId=... to deactivate and remove the consent.
// consentId is the consent previously authorized for AutoDebit; it can be retrieved from payment method details enquiry.
// Authorization uses JWT; for DELETE with no body, the "data" claim is omitted per DuitNow Pay API Authentication.
func TerminateConsent(cfg ClientConfig, consentId string) (*TerminateConsentResponse, int, error) {
	if consentId == "" {
		return nil, 0, fmt.Errorf("consentId is required")
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
		JTI:        strings.TrimSpace(consentId),
		Key:        cfg.JWTKey,
		Data:       nil, // no body for DELETE; do not include "data" claim
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	rawURL := baseURL + "/v1/bw/consent"
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, 0, fmt.Errorf("parse URL: %w", err)
	}
	q := u.Query()
	q.Set("consentId", consentId)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[terminate_consent] --- Outgoing request to PayNet ---")
	log.Printf("[terminate_consent] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[terminate_consent] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[terminate_consent]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[terminate_consent] -----------------------------------------")

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

	log.Printf("[terminate_consent] --- Response from PayNet ---")
	log.Printf("[terminate_consent] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[terminate_consent] Body:\n%s", string(respBody))
	log.Printf("[terminate_consent] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var termResp TerminateConsentResponse
	if err := json.Unmarshal(respBody, &termResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &termResp, resp.StatusCode, nil
}

// SampleConsentId returns a sample consentId for trigger/testing (as in the API spec example).
func SampleConsentId() string {
	return "M00002010012700006"
}
