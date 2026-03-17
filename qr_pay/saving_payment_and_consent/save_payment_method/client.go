package save_payment_method

import (
	"bytes"
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
// https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/save-payment-method

// ClientConfig holds configuration for the PayNet DuitNow Pay save payment method client.
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

// SampleRequest returns a sample SavePaymentMethodRequest (dataType "02") per PayNet docs.
func SampleRequest() SavePaymentMethodRequest {
	return SavePaymentMethodRequest{
		// UPDATE HERE
		// DataType:      "02",
		DataType:      "01",
		CheckoutID:    "a7e2ed2a-b088-4495-8cf4-88da08f644f2",
		SourceOfFunds: []string{"01"},
		MerchantName:  "Shop Name Sdn Bhd.",
		MerchantRefID: "ref12345678",
		Merchant:      Merchant{ProductID: "P00000201"},
		Customer: Customer{
			Name:               "Walter Mitty",
			IdentityValidation: "00",
			IdentificationType: "05",
			Identification:     "+60123456789",
		},
		Consent: Consent{
			MaxAmount:     "500.00",
			EffectiveDate: "2024-01-24",
			ExpiryDate:    "2024-04-24",
			Frequency:     "01", // Unlimited
		},
		Language: "en",
	}
}

// CreateSavePaymentMethod sends a POST request to PayNet's /v1/payment/intent with dataType "02" and returns the response.
// Authorization uses JWT with the request body as the "data" claim per DuitNow Pay API Authentication.
func CreateSavePaymentMethod(cfg ClientConfig, req SavePaymentMethodRequest) (*SavePaymentMethodResponse, int, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	privateKey, err := jwt_generation.LoadDefaultPrivateKey()
	if err != nil {
		return nil, 0, fmt.Errorf("load private key for JWT: %w", err)
	}

	dataClaim := json.RawMessage(bodyBytes)
	jti := strings.TrimSpace(req.CheckoutID)
	if jti == "" {
		jti = "550e8400-e29b-41d4-a716-446655440000"
	}

	token, err := jwt_generation.GenerateJWT(jwt_generation.GenerateOptions{
		PrivateKey: privateKey,
		Algorithm:  jwt_generation.RS256,
		Issuer:     cfg.JWTIssuer,
		Subject:    cfg.JWTSubject,
		JTI:        jti,
		Key:        cfg.JWTKey,
		Data:       dataClaim,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("generate JWT: %w", err)
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	url := baseURL + "/v1/payment/intent"

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	common_header.ApplyToRequest(httpReq, common_header.Default())

	log.Printf("[save_payment_method] --- Outgoing request to PayNet ---")
	log.Printf("[save_payment_method] Method: %s URL: %s", httpReq.Method, httpReq.URL.String())
	log.Printf("[save_payment_method] Headers:")
	for k, v := range httpReq.Header {
		log.Printf("[save_payment_method]   %s: %s", k, strings.Join(v, ", "))
	}
	bodyIndent, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[save_payment_method] Body:\n%s", string(bodyIndent))
	log.Printf("[save_payment_method] -----------------------------------------")

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

	log.Printf("[save_payment_method] --- Response from PayNet ---")
	log.Printf("[save_payment_method] Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[save_payment_method] Body:\n%s", string(respBody))
	log.Printf("[save_payment_method] -----------------------------------------")

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var intentResp SavePaymentMethodResponse
	if err := json.Unmarshal(respBody, &intentResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &intentResp, resp.StatusCode, nil
}
