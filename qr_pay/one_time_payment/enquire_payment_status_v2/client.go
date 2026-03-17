package enquire_payment_status_v2

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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-status

// PaymentMethodOneTime is the payment method code for one-time payment (DuitNow Online Banking / Wallets).
const PaymentMethodOneTime = "01"

// PaymentMethodSaved is the payment method code for saved payment method.
const PaymentMethodSaved = "02"

// ClientConfig holds configuration for the PayNet DuitNow Pay Enquire Payment Status v2 client.
type ClientConfig struct {
	// BaseURL is the PayNet DuitNow Pay API base URL (e.g. https://certification.api.developer.inet.paynet.my/pay-guard). Path /v2/bw/checkout-status is appended.
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

// EnquirePaymentStatusV2 sends a GET request to PayNet's /v2/bw/checkout-status?paymentMethod=...&checkoutId=... to retrieve the payment status.
// paymentMethod: 01 - One-time payment, 02 - Saved payment method.
// checkoutId is the unique external identifier (UUID v4) provided by the acquirer when initiating a payment intent.
// API has a rate limit: can only be called once every 30 seconds per transaction.
// Authorization uses JWT; for GET with no body, the "data" claim is omitted per DuitNow Pay API Authentication.
func EnquirePaymentStatusV2(cfg ClientConfig, paymentMethod, checkoutId string) (*EnquirePaymentStatusV2Response, int, error) {
	if paymentMethod == "" {
		paymentMethod = PaymentMethodOneTime
	}
	if checkoutId == "" {
		checkoutId = "00000000-0000-0000-0000-000000000001"
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
		JTI:        "550e8400-e29b-41d4-a716-446655440002",
		Key:        cfg.JWTKey,
		Data:       nil, // no body for GET; do not include "data" claim
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	rawURL := baseURL + "/v2/bw/checkout-status"
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, 0, fmt.Errorf("parse URL: %w", err)
	}
	q := u.Query()
	q.Set("paymentMethod", paymentMethod)
	q.Set("checkoutId", checkoutId)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[enquire_payment_status_v2] --- Outgoing request to PayNet ---")
	log.Printf("[enquire_payment_status_v2] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[enquire_payment_status_v2] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[enquire_payment_status_v2]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[enquire_payment_status_v2] -----------------------------------------")

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

	log.Printf("[enquire_payment_status_v2] --- Response from PayNet ---")
	log.Printf("[enquire_payment_status_v2] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[enquire_payment_status_v2] Body:\n%s", string(respBody))
	log.Printf("[enquire_payment_status_v2] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResp EnquirePaymentStatusV2Response
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &statusResp, resp.StatusCode, nil
}

// SampleCheckoutId returns a sample checkoutId for trigger/testing (as in the API spec example).
func SampleCheckoutId() string {
	return "00000000-0000-0000-0000-000000000001"
}
