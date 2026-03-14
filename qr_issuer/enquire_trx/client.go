package enquire_trx

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

// ClientConfig holds configuration for the PayNet Issuer transactions enquire client.
type ClientConfig struct {
	BaseURL          string
	ClientID         string
	ApiVersion       string
	JWSIssuer        string
	JWSCredentialKey string
}

// DefaultClientConfig returns a config with default JWS issuer/credential key (same pattern as account_enquire_xc).
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		BaseURL:          "https://certification.api.developer.inet.paynet.my/v1/picasso-guard",
		ClientID:         "MBBEMYKL",
		ApiVersion:       "v3",
		JWSIssuer:        "MBBEMYKL",
		JWSCredentialKey: "64feb830",
	}
}

// SampleRequest returns a sample TransactionEnquiryRequest for testing.
// BusinessMessageId format: YYYYMMDDBBBBBBBBXXXOCCSSSSSSSS (e.g. 630 = transaction enquire).
func SampleRequest() TransactionEnquiryRequest {
	return TransactionEnquiryRequest{
		AppHeader: TrxEnquiryAppHeader{
			BusinessMessageId: "20250709TESTMYKL630ORM00000060",
			CreationDateTime:  "2025-07-09T12:31:56.170+08:00",
			TransactionId:     "20250709TESTMYKL030OQR00057310",
		},
		DebtorAgent:   Agent{Id: "TESTMYKL"},
		CreditorAgent: Agent{Id: "TST1MYKL"},
	}
}

// EnquireTrx sends a POST request to PayNet's /v3/transactions/enquire and returns the response.
// Authorization header is built using JWS over the request body (same approach as account_enquire_xc).
func EnquireTrx(cfg ClientConfig, req TransactionEnquiryRequest) (*TransactionEnquiryResponse, int, error) {
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
		issuer = "RPPEMYKL"
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
	url := baseURL + "/v3/transactions/enquire"

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

	log.Printf("[enquire_trx] --- Outgoing request to external API ---")
	log.Printf("[enquire_trx] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[enquire_trx] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		httpReq.Header.Get("X-Client-Id"),
		httpReq.Header.Get("X-Api-Version"),
		httpReq.Header.Get("x-business-message-id"))
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[enquire_trx] Body:\n%s", string(bodyIndent))
	log.Printf("[enquire_trx] -----------------------------------------")

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

	log.Printf("[enquire_trx] --- Response from external API ---")
	log.Printf("[enquire_trx] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[enquire_trx] Response Body:\n%s", string(respBody))
	log.Printf("[enquire_trx] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var trxResp TransactionEnquiryResponse
	if err := json.Unmarshal(respBody, &trxResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &trxResp, resp.StatusCode, nil
}
