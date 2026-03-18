package payment_method_status

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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-method-status
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-details

// ClientConfig holds configuration for the PayNet DuitNow Pay Enquire Payment Method Status client.
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

// GetPaymentMethodStatus sends a GET request to PayNet's /v1/bw/consent/request?checkoutId=... to retrieve consent status.
// checkoutId is the unique external identifier (UUID v4) provided by the acquirer when initiating consent.
// API has a rate limit: can only be called once every 30 seconds per transaction.
// Authorization uses JWT; for GET with no body, the "data" claim is omitted per DuitNow Pay API Authentication.
func GetPaymentMethodStatus(cfg ClientConfig, checkoutId string) (*PaymentMethodStatusResponse, int, error) {
	if checkoutId == "" {
		return nil, 0, fmt.Errorf("checkoutId is required")
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
		JTI:        strings.TrimSpace(checkoutId),
		Key:        cfg.JWTKey,
		Data:       nil, // no body for GET; do not include "data" claim
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	rawURL := baseURL + "/v1/bw/consent/request"
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, 0, fmt.Errorf("parse URL: %w", err)
	}
	q := u.Query()
	q.Set("checkoutId", checkoutId)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[payment_method_status] --- Outgoing request to PayNet ---")
	log.Printf("[payment_method_status] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[payment_method_status] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[payment_method_status]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[payment_method_status] -----------------------------------------")

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

	log.Printf("[payment_method_status] --- Response from PayNet ---")
	log.Printf("[payment_method_status] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[payment_method_status] Body:\n%s", string(respBody))
	log.Printf("[payment_method_status] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResp PaymentMethodStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &statusResp, resp.StatusCode, nil
}

// SampleCheckoutId returns a sample checkoutId for trigger/testing (as in the API spec example).
func SampleCheckoutId() string {
	return "a7e2ed2a-b088-4495-8cf4-88da08f644f2"
}

// GetPaymentMethodDetails sends a GET request to PayNet's /v1/bw/consent?consentId=... to retrieve consent details.
// consentId is the consent authorized for AutoDebit, received from the Update Consent Details webhook.
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-details
func GetPaymentMethodDetails(cfg ClientConfig, consentId string) (*PaymentMethodDetailsResponse, int, error) {
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
		Data:       nil,
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

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[payment_method_status] --- Outgoing request to PayNet (Enquire Payment Method Details) ---")
	log.Printf("[payment_method_status] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[payment_method_status] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[payment_method_status]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[payment_method_status] -----------------------------------------")

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

	log.Printf("[payment_method_status] --- Response from PayNet ---")
	log.Printf("[payment_method_status] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[payment_method_status] Body:\n%s", string(respBody))
	log.Printf("[payment_method_status] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var detailsResp PaymentMethodDetailsResponse
	if err := json.Unmarshal(respBody, &detailsResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &detailsResp, resp.StatusCode, nil
}

// SampleConsentId returns a sample consentId for trigger/testing (as in the payment-details API spec example).
func SampleConsentId() string {
	return "M00002010012700006"
}
