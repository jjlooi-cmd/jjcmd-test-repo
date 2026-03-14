package payments_reverse

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

// ClientConfig holds configuration for the PayNet DuitNow Reversal client (POST /v3/payments/reverse).
type ClientConfig struct {
	// BaseURL is the PayNet API base URL. Path /v3/payments/reverse is appended.
	BaseURL string
	// ClientID is the X-Client-Id header value (from onboarding).
	ClientID string
	// ApiVersion is the X-Api-Version header value (e.g. v3).
	ApiVersion string
	// JWSIssuer is the issuer claim (e.g. BIC code) used to sign the request.
	JWSIssuer string
	// JWSCredentialKey is the JWT Credential Key assigned during onboarding.
	JWSCredentialKey string
	// GPSCoordinates optional; location of sender (decimal degree). Spec: x-gps-coordinates.
	GPSCoordinates string
	// IPAddress is required by the API; IP where transaction originated (IPv4 or IPv6). Spec: x-ip-address.
	IPAddress string
}

// DefaultClientConfig returns a config with default JWS issuer/credential key (same pattern as account_enquire_xc).
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		BaseURL:          "https://certification.api.developer.inet.paynet.my/v1/picasso-guard",
		ClientID:         "MBBEMYKL",
		ApiVersion:       "v3",
		JWSIssuer:        "MBBEMYKL",
		JWSCredentialKey: "64feb830",
		IPAddress:        "74.220.48.246",
	}
}

// SampleRequest returns a sample PaymentReverseRequest for testing.
// BusinessMessageId format: YYYYMMDD(8) + BIC(8) + XXX(3) + O(1) + CC(2) + SSSSSSSS(8). Transaction type 011 for DuitNow Transfer/DuitNowQR.
func SampleRequest() PaymentReverseRequest {
	bmid := "20260314MBBEMYKL011OQR95535834"
	return PaymentReverseRequest{
		AppHeader: ReverseAppHeader{
			EndToEndId:        bmid,
			TransactionId:     "20260314MBBEMYKL030OQR00057310",
			BusinessMessageId: bmid,
			CreationDateTime:  "2026-03-14T17:00:00.000+08:00",
		},
		InterbankSettlementAmount: InterbankSettlementAmount{
			Value:    10.00,
			Currency: "MYR",
		},
		Debtor: ReverseParty{
			Name: "Jane Smith",
			Type: "RET",
		},
		DebtorAccount: ReverseDebtorAccount{
			Id:   "0123456789",
			Type: "DEFAULT",
		},
		DebtorAgent: ReverseAgent{
			Id: "111555",
		},
		Creditor: ReverseParty{
			Name: "Sample Debtor Name",
			Type: "RET",
		},
		CreditorAccount: ReverseCreditorAccount{
			Id:   "1234567890",
			Type: "CURRENT",
		},
		CreditorAgent: ReverseAgent{
			Id: "MBBEMYKL",
		},
		PaymentDescription:    "Reversal of QR payment",
		AcceptedSourceOfFunds: []string{"CASA"},
	}
}

// Reverse sends a POST request to PayNet's v3/payments/reverse and returns the response.
// Authorization is JWS over the request body (same approach as qr_issuer/account_enquire_xc).
// Ref: document (4).yaml — DuitNow Reversal (Acquirer initiates reversal to PayNet).
func Reverse(cfg ClientConfig, req PaymentReverseRequest) (*PaymentReverseResponse, int, error) {
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
	url := baseURL + "/v3/payments/reverse"

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
	if cfg.GPSCoordinates != "" {
		httpReq.Header.Set("x-gps-coordinates", cfg.GPSCoordinates)
	}
	if cfg.IPAddress != "" {
		httpReq.Header.Set("x-ip-address", cfg.IPAddress)
	}

	log.Printf("[payments_reverse] --- Outgoing request to external API ---")
	log.Printf("[payments_reverse] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[payments_reverse] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s x-ip-address=%s",
		httpReq.Header.Get("X-Client-Id"),
		httpReq.Header.Get("X-Api-Version"),
		httpReq.Header.Get("x-business-message-id"),
		httpReq.Header.Get("x-ip-address"))
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[payments_reverse] Body:\n%s", string(bodyIndent))
	log.Printf("[payments_reverse] -----------------------------------------")

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

	log.Printf("[payments_reverse] --- Response from external API ---")
	log.Printf("[payments_reverse] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[payments_reverse] Response Body:\n%s", string(respBody))
	log.Printf("[payments_reverse] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var reverseResp PaymentReverseResponse
	if err := json.Unmarshal(respBody, &reverseResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &reverseResp, resp.StatusCode, nil
}
