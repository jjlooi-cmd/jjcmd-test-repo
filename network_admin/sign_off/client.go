package sign_off

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"example.com/sample-repo/jws_generation"
)

// ClientConfig holds configuration for the PayNet Network Admin sign-off client.
type ClientConfig struct {
	// BaseURL is the PayNet API base URL (e.g. https://api.paynet.my). Path /v3/admin/sign-off is appended.
	BaseURL string
	// ClientID is the X-Client-Id header value (Bank Identification Code).
	ClientID string
	// ApiVersion is the X-Api-Version header value (e.g. 1.0.0).
	ApiVersion string
	// JWSIssuer is the issuer claim (e.g. BIC code) used to sign the request.
	JWSIssuer string
	// JWSCredentialKey is the JWT Credential Key assigned during onboarding.
	JWSCredentialKey string
}

// DefaultClientConfig returns a config with default values for testing.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		BaseURL:          "https://certification.api.developer.inet.paynet.my/v1/picasso-guard",
		ClientID:         "MBBEMYKL",
		ApiVersion:       "1.0.0",
		JWSIssuer:        "MBBEMYKL",
		JWSCredentialKey: "64feb830",
	}
}

// SampleRequest returns a sample SystemAdminRequest for POST /v3/admin/sign-off (transaction code 000).
// BusinessMessageId format: YYYYMMDD(8) + BIC(8) + TxnCode 000(3) + Originator(1) + Channel(2) + Sequence(8).
func SampleRequest() SystemAdminRequest {
	return SystemAdminRequest{
		AppHeader: SystemAdminRequestAppHeader{
			BusinessMessageId: "20260314MBBEMYKL000ORB00018772",
			CreationDateTime:  "2026-03-14T18:12:43.903+08:00",
		},
	}
}

// SignOff sends a POST request to PayNet's /v3/admin/sign-off to disconnect from RPP (RPP transaction code 000).
// Authorization header is built using JWS over the request body.
// Ref: document (6).yaml — POST /v3/admin/sign-off
func SignOff(cfg ClientConfig, req SystemAdminRequest) (*SystemAdminResponse, int, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWS: %w", err)
	}

	issuer := strings.TrimSpace(cfg.JWSIssuer)
	if issuer == "" {
		issuer = "MBBEMYKL"
	}
	credKey := strings.TrimSpace(cfg.JWSCredentialKey)
	if credKey == "" {
		credKey = "64feb830"
	}

	token, err := jws_generation.GenerateJWS(jws_generation.GenerateOptions{
		PrivateKey:        privateKey,
		Algorithm:         jws_generation.RS512,
		Issuer:            issuer,
		BusinessMessageID: req.AppHeader.BusinessMessageId,
		CredentialKey:     credKey,
		PayloadForHash:    bodyBytes,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWS: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	url := baseURL + "/v3/admin/sign-off"

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	if cfg.ClientID != "" {
		httpReq.Header.Set("X-Client-Id", cfg.ClientID)
	}
	if cfg.ApiVersion != "" {
		httpReq.Header.Set("X-Api-Version", cfg.ApiVersion)
	}
	if req.AppHeader.BusinessMessageId != "" {
		httpReq.Header.Set("x-business-message-id", req.AppHeader.BusinessMessageId)
	}

	log.Printf("[sign_off] --- Outgoing request to /v3/admin/sign-off ---")
	log.Printf("[sign_off] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[sign_off] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		httpReq.Header.Get("X-Client-Id"),
		httpReq.Header.Get("X-Api-Version"),
		httpReq.Header.Get("x-business-message-id"))
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[sign_off] Body:\n%s", string(bodyIndent))
	log.Printf("[sign_off] -----------------------------------------")

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

	log.Printf("[sign_off] --- Response from /v3/admin/sign-off ---")
	log.Printf("[sign_off] Status: %d %s", resp.StatusCode, resp.Status)
	for k, v := range resp.Header {
		log.Printf("[sign_off]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[sign_off] Body:\n%s", string(respBody))
	log.Printf("[sign_off] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var adminResp SystemAdminResponse
	if err := json.Unmarshal(respBody, &adminResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &adminResp, resp.StatusCode, nil
}
