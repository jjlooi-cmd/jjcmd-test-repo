package account_enquire_xc

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

// ClientConfig holds configuration for the PayNet Issuer account enquire-xc client.
type ClientConfig struct {
	// BaseURL is the PayNet API base URL (e.g. https://api.paynet.my). Path /v3/accounts/enquire-xc is appended.
	BaseURL string
	// ClientID is the X-Client-Id header value (from onboarding).
	ClientID string
	// ApiVersion is the X-Api-Version header value (e.g. v3).
	ApiVersion string
	// JWSIssuer is the issuer claim (e.g. issuer BIC code) used to sign the request.
	JWSIssuer string
	// JWSCredentialKey is the JWT Credential Key assigned during onboarding.
	JWSCredentialKey string
}

// DefaultClientConfig returns a config with default JWS issuer/credential key (same pattern as qr_acquirer).
// Override BaseURL, ClientID, and ApiVersion for your environment.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		// BaseURL:          "https://api.paynet.my",
		BaseURL:          "https://certification.api.developer.inet.paynet.my/v1/picasso-guard",
		ClientID:         "MBBEMYKL",
		ApiVersion:       "v3",
		JWSIssuer:        "MBBEMYKL",
		JWSCredentialKey: "64feb830",
	}
}

// SampleRequest returns a sample EnquireRequest with values filled in for testing.
// Use with EnquireXC: resp, code, err := EnquireXC(DefaultClientConfig(), SampleRequest()).
// BusinessMessageId format: YYYYMMDD(8) + BIC(8) + TxnCode(3) + Originator(1) + Channel(2) + Sequence(8).
func SampleRequest() EnquireRequest {
	return EnquireRequest{
		AppHeader: AppHeader{
			EndToEndId:        "20260313MBBEMYKL520OQR68495070",
			BusinessMessageId: "20260313MBBEMYKL520HQR95535833",
			CreationDateTime: "2026-03-13T10:30:00.000+08:00",
		},
		Debtor: Party{
			Id:   "DEBTOR001",
			Name: "Sample Debtor Name",
		},
		DebtorAccount: DebtorAccount{
			Id:                "1234567890",
			Type:              "CURRENT",
			ResidentStatus:    "RESIDENT",
			ProductType:       "ISLAMIC",
			ShariaCompliance:  "YES",
			AccountHolderType: "SINGLE",
		},
		DebtorAgent: Agent{
			Id: "MBBEMYKL",
		},
		CreditorAgent: Agent{
			Id: "MBBEMYKL",
		},
		CreditorAccount: CreditorAccount{
			Id:   "123456789",
			Type: "DEFAULT",
		},
		QR: QR{
			Code: "00020201021226420014A00000061500010106111555021001234567895204000153034585802MY5916Kedai CU Sdn Bhd6012Kuala Lumpur6304D9D6",
		},
	}
}

// EnquireXC sends a POST request to PayNet's v3/accounts/enquire-xc (Issuer) and returns the response.
// Authorization header is built using JWS over the request body, following the same approach as qr_acquirer/account_enquire_xc.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/issuer/domestic#/paths/v3-accounts-enquire-xc/post
func EnquireXC(cfg ClientConfig, req EnquireRequest) (*EnquireResponse, int, error) {
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
	url := baseURL + "/v3/accounts/enquire-xc"

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

	// Log outgoing request before sending
	log.Printf("[account_enquire_xc] --- Outgoing request to external API ---")
	log.Printf("[account_enquire_xc] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[account_enquire_xc] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		httpReq.Header.Get("X-Client-Id"),
		httpReq.Header.Get("X-Api-Version"),
		httpReq.Header.Get("x-business-message-id"))
	log.Printf("[account_enquire_xc] Authorization (Bearer token): %s", httpReq.Header.Get("Authorization"))
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[account_enquire_xc] Body:\n%s", string(bodyIndent))
	log.Printf("[account_enquire_xc] -----------------------------------------")

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

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var enqResp EnquireResponse
	if err := json.Unmarshal(respBody, &enqResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &enqResp, resp.StatusCode, nil
}
