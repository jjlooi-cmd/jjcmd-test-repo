package webhook_update_consent_details

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler implements POST /pg-router/v1/duitnowpay/consent-notification for the PayNet Webhook: Update Consent Details.
// This webhook updates the acquirer when a save payment method is initiated. It returns the consentId with the status.
// Acquirer shall provide an acknowledgement back to API Gateway.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-consent#webhook-update-consent-details
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[webhook_update_consent_details] Webhook triggered: %s %s", r.Method, r.URL.String())
	if len(r.Header) > 0 {
		log.Printf("[webhook_update_consent_details] Headers:")
		for k, v := range r.Header {
			log.Printf("[webhook_update_consent_details]   %s: %s", k, strings.Join(v, ", "))
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[webhook_update_consent_details] failed to read body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read request body"})
		return
	}
	defer r.Body.Close()
	if len(body) > 0 {
		log.Printf("[webhook_update_consent_details] Body (raw): %s", string(body))
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "POST required"})
		return
	}

	var req UpdateConsentDetailsRequest
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
		log.Printf("[webhook_update_consent_details] invalid JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Request body must be valid JSON"})
		return
	}

	log.Printf("[webhook_update_consent_details] --- Incoming request ---")
	log.Printf("[webhook_update_consent_details] Method: %s URL: %s", r.Method, r.URL.String())
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[webhook_update_consent_details] Body:\n%s", string(bodyBytes))
	log.Printf("[webhook_update_consent_details] ------------------------")

	if strings.TrimSpace(req.CheckoutId) == "" {
		log.Printf("[webhook_update_consent_details] missing checkoutId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "checkoutId is required"})
		return
	}
	if strings.TrimSpace(req.EndToEndId) == "" {
		log.Printf("[webhook_update_consent_details] missing endToEndId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "endToEndId is required"})
		return
	}
	if strings.TrimSpace(req.Issuer) == "" {
		log.Printf("[webhook_update_consent_details] missing issuer")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "issuer is required"})
		return
	}
	if strings.TrimSpace(req.ConsentStatus.ConsentId) == "" {
		log.Printf("[webhook_update_consent_details] missing consentStatus.consentId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "consentStatus.consentId is required"})
		return
	}
	if strings.TrimSpace(req.ConsentStatus.Code) == "" {
		log.Printf("[webhook_update_consent_details] missing consentStatus.code")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "consentStatus.code is required"})
		return
	}
	if strings.TrimSpace(req.ConsentStatus.Message) == "" {
		log.Printf("[webhook_update_consent_details] missing consentStatus.message")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "consentStatus.message is required"})
		return
	}

	// Acquirer business logic: e.g. persist consent, update saved payment method state, notify user.
	// Replace with your storage or downstream processing.
	_ = req

	log.Printf("[webhook_update_consent_details] Consent details received: checkoutId=%s endToEndId=%s consentId=%s code=%s issuer=%s",
		req.CheckoutId, req.EndToEndId, req.ConsentStatus.ConsentId, req.ConsentStatus.Code, req.Issuer)

	// Acquirer shall provide an acknowledgement back to API Gateway (per API spec).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "OK"})
}
