package get_bank_list

import (
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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/get-bank-list

// ClientConfig holds configuration for the PayNet DuitNow Pay Get Bank List client.
type ClientConfig struct {
	// BaseURL is the PayNet DuitNow Pay API base URL (e.g. https://certification.api.developer.inet.paynet.my/pay-guard). Path /v2/bw/banks is appended.
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

// GetBankList sends a GET request to PayNet's /v2/bw/banks and returns the bank list (retail and corporate).
// Participants are recommended to call this API on a 30-minute interval.
// Authorization uses JWT; for GET with no body, the "data" claim is an empty object per DuitNow Pay API Authentication.
func GetBankList(cfg ClientConfig) (*GetBankListResponse, int, error) {
	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

	// GET has no body; omit "data" claim so the JWT matches PayNet's expectation for body-less requests.
	token, err := jwt_generation.GenerateJWT(jwt_generation.GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  jwt_generation.RS256,
		Issuer:     cfg.JWTIssuer,
		Subject:    cfg.JWTSubject,
		JTI:        "550e8400-e29b-41d4-a716-446655440000",
		Key:        cfg.JWTKey,
		Data:       nil, // no body for GET; do not include "data" claim
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	url := baseURL + "/v2/bw/banks"

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[get_bank_list] --- Outgoing request to PayNet ---")
	log.Printf("[get_bank_list] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[get_bank_list] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[get_bank_list]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[get_bank_list] -----------------------------------------")

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

	log.Printf("[get_bank_list] --- Response from PayNet ---")
	log.Printf("[get_bank_list] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[get_bank_list] Body:\n%s", string(respBody))
	log.Printf("[get_bank_list] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var bankListResp GetBankListResponse
	if err := json.Unmarshal(respBody, &bankListResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &bankListResp, resp.StatusCode, nil
}
