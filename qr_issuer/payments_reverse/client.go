package payments_reverse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/sample-repo/jws_generation"
)

// ClientConfig holds configuration for the PayNet DuitNow Reversal client (outbound call to PayNet).
// Path /v3/payments/reverse is appended to BaseURL.
type ClientConfig struct {
	BaseURL          string
	ClientID         string
	ApiVersion       string
	JWSIssuer        string
	JWSCredentialKey string
	GPSCoordinates   string
	IPAddress        string
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

// SampleRequest returns a sample PaymentReverseWebhookRequest for testing.
// BusinessMessageId: use reversal transaction type (e.g. 011/012); format YYYYMMDDBBBBBBBBXXXOCCSSSSSSSS.
func SampleRequest() PaymentReverseWebhookRequest {
	bmid := "20260314MBBEMYKL011RQR95535834"
	origTxnId := "20260314MBBEMYKL030OQR00057310"
	return PaymentReverseWebhookRequest{
		SettlementCycleNumber:   "001",
		InterbankSettlementDate: time.Now().Format("2006-01-02"),
		AppHeader: PaymentReverseAppHeader{
			EndToEndId:        bmid,
			TransactionId:     origTxnId,
			BusinessMessageId: bmid,
			CreationDateTime:  time.Now().Format(time.RFC3339),
		},
		InterbankSettlementAmount: InterbankSettlementAmount{
			Value:    10.00,
			Currency: "MYR",
		},
		Debtor: PaymentReverseParty{
			Name: "Jane Smith",
			Type: "RET",
		},
		DebtorAccount: PaymentReverseAccount{
			Id:   "0123456789",
			Type: "SAVINGS",
		},
		DebtorAgent: PaymentReverseAgent{
			Id: "111555",
		},
		Creditor: PaymentReverseParty{
			Name: "Sample Debtor Name",
			Type: "RET",
		},
		CreditorAccount: PaymentReverseAccount{
			Id:   "1234567890",
			Type: "CURRENT",
		},
		CreditorAgent: PaymentReverseAgent{
			Id: "MBBEMYKL",
		},
		PaymentDescription: "Reversal of QR payment",
	}
}

// ReverseXC sends a POST request to PayNet's v3/payments/reverse and returns the response.
// Authorization is JWS over the request body. On 200 returns PaymentReverseResponse; on 400 returns ErrorResponse in err or a separate out param if needed.
// Ref: document (3).yaml - DuitNow Reversal (outbound call to submit reversal).
func ReverseXC(cfg ClientConfig, req PaymentReverseWebhookRequest) (*PaymentReverseResponse, int, http.Header, error) {
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
	url := baseURL + "/v3/payments/reverse"

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

	log.Printf("[payments_reverse] --- Outgoing request to external API ---")
	log.Printf("[payments_reverse] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[payments_reverse] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		httpReq.Header.Get("X-Client-Id"),
		httpReq.Header.Get("X-Api-Version"),
		httpReq.Header.Get("x-business-message-id"))
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[payments_reverse] Body:\n%s", string(bodyIndent))
	log.Printf("[payments_reverse] -----------------------------------------")

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

	respHeader := make(http.Header)
	for k, v := range resp.Header {
		respHeader[k] = v
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, respHeader, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var reverseResp PaymentReverseResponse
	if err := json.Unmarshal(respBody, &reverseResp); err != nil {
		return nil, resp.StatusCode, respHeader, fmt.Errorf("decode response: %w", err)
	}

	return &reverseResp, resp.StatusCode, respHeader, nil
}
