package payments_transfer_xc

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

// ClientConfig holds configuration for the PayNet Issuer payments transfer-xc client.
type ClientConfig struct {
	// BaseURL is the PayNet API base URL (e.g. https://api.paynet.my). Path /v3/payments/transfer-xc is appended.
	BaseURL string
	// ClientID is the X-Client-Id header value (from onboarding).
	ClientID string
	// ApiVersion is the X-Api-Version header value (e.g. v3).
	ApiVersion string
	// JWSIssuer is the issuer claim (e.g. issuer BIC code) used to sign the request.
	JWSIssuer string
	// JWSCredentialKey is the JWT Credential Key assigned during onboarding.
	JWSCredentialKey string
	// GPSCoordinates optional; location of sender (decimal degree). Spec: x-gps-coordinates.
	GPSCoordinates string
	// IPAddress optional; IP where transaction originated (IPv4 or IPv6). Spec: x-ip-address.
	IPAddress string
}

// DefaultClientConfig returns a config with default JWS issuer/credential key (same pattern as account_enquire_xc).
// Override BaseURL, ClientID, and ApiVersion for your environment.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		BaseURL:          "https://certification.api.developer.inet.paynet.my/v1/picasso-guard",
		ClientID:         "MBBEMYKL",
		ApiVersion:       "v3",
		JWSIssuer:        "MBBEMYKL",
		JWSCredentialKey: "64feb830",
	}
}

// SampleRequest returns a sample TransferRequest with values filled in for testing.
// Use with TransferXC: resp, code, headers, err := TransferXC(DefaultClientConfig(), SampleRequest()).
// BusinessMessageId format: YYYYMMDD(8) + BIC(8) + TxnCode(3) + Originator(1) + Channel(2) + Sequence(8).
// Spec: use QR Enquiry (520) BMID with transaction code 030 (POS) or 040 (P2P); creditor.name from enquiry response.
func SampleRequest() TransferRequest {
	bmid := "20260314MBBEMYKL520OQR95535833"
	return TransferRequest{
		AppHeader: AppHeader{
			EndToEndId: bmid,
			// TransactionId:     bmid,
			TransactionId:     "20260314MBBEMYKL030OQR00057310",
			BusinessMessageId: bmid,
			CreationDateTime:  "2026-03-14T00:30:00.000+08:00",
		},
		InterbankSettlementAmount: InterbankSettlementAmount{
			Value:    DecimalAmount(10.00),
			Currency: "MYR",
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
		Creditor: Creditor{
			Name: "Jane Smith",
		},
		CreditorAccount: CreditorAccount{
			Id: "0123456789",
		},
		CreditorAgent: Agent{
			Id: "111555",
		},
		RecipientReference: "QR Payment",
		PaymentDescription: "Lunch at Nasi Lemak Shop",
		QR: QR{
			Code:                  "00020201021226420014A00000061500010106111555021001234567895204000153034585802MY5916Kedai CU Sdn Bhd6012Kuala Lumpur6304D9D6",
			Category:              "POINT_OF_SALES",
			AcceptedSourceOfFunds: []string{"CASA"},
			PromoCode:             "PROMO123",
		},
	}
}

// TransferXC sends a POST request to PayNet's v3/payments/transfer-xc (Issuer) and returns the response.
// Authorization header is built using JWS over the request body, following the same approach as qr_issuer/account_enquire_xc.
// Returns the parsed response body, HTTP status code, response headers (clone), and an error.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/issuer/domestic#/paths/v3-payments-transfer-xc/post
func TransferXC(cfg ClientConfig, req TransferRequest) (*TransferResponse, int, http.Header, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("marshal request: %w", err)
	}

	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("load private key for JWS: %w", err)
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
		return nil, 0, nil, fmt.Errorf("generate JWS: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	url := baseURL + "/v3/payments/transfer-xc"

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("create request: %w", err)
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

	log.Printf("[payments_transfer_xc] --- Outgoing request to external API ---")
	log.Printf("[payments_transfer_xc] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[payments_transfer_xc] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		httpReq.Header.Get("X-Client-Id"),
		httpReq.Header.Get("X-Api-Version"),
		httpReq.Header.Get("x-business-message-id"))
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[payments_transfer_xc] Body:\n%s", string(bodyIndent))
	log.Printf("[payments_transfer_xc] -----------------------------------------")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}

	// Clone response headers for caller (resp may be closed)
	respHeader := make(http.Header)
	for k, v := range resp.Header {
		respHeader[k] = v
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, respHeader, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var transferResp TransferResponse
	if err := json.Unmarshal(respBody, &transferResp); err != nil {
		return nil, resp.StatusCode, respHeader, fmt.Errorf("decode response: %w", err)
	}

	return &transferResp, resp.StatusCode, respHeader, nil
}
