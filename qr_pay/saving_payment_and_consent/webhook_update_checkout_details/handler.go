package webhook_update_checkout_details

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler implements POST /pg-router/v1/duitnowpay/consent-notification for the PayNet Webhook: Update Checkout Details (consent flow).
// This webhook maps the endToEndId to the checkoutId so the acquirer can relate the endToEndId in the redirect URL
// back to the checkoutId when the issuer redirects with only the endToEndId. Acquirer shall provide an acknowledgement back to API Gateway.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-consent#webhook-update-checkout-details
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[webhook_update_checkout_details] Webhook triggered: %s %s", r.Method, r.URL.String())
	if len(r.Header) > 0 {
		log.Printf("[webhook_update_checkout_details] Headers:")
		for k, v := range r.Header {
			log.Printf("[webhook_update_checkout_details]   %s: %s", k, strings.Join(v, ", "))
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[webhook_update_checkout_details] failed to read body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read request body"})
		return
	}
	defer r.Body.Close()
	if len(body) > 0 {
		log.Printf("[webhook_update_checkout_details] Body (raw): %s", string(body))
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "POST required"})
		return
	}

	var req UpdateCheckoutDetailsRequest
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
		log.Printf("[webhook_update_checkout_details] invalid JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Request body must be valid JSON"})
		return
	}

	log.Printf("[webhook_update_checkout_details] --- Incoming request ---")
	log.Printf("[webhook_update_checkout_details] Method: %s URL: %s", r.Method, r.URL.String())
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[webhook_update_checkout_details] Body:\n%s", string(bodyBytes))
	log.Printf("[webhook_update_checkout_details] ------------------------")

	if strings.TrimSpace(req.CheckoutId) == "" {
		log.Printf("[webhook_update_checkout_details] missing checkoutId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "checkoutId is required"})
		return
	}
	if strings.TrimSpace(req.ConsentEndToEndId) == "" {
		log.Printf("[webhook_update_checkout_details] missing consentEndToEndId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "consentEndToEndId is required"})
		return
	}
	if strings.TrimSpace(req.ConsentId) == "" {
		log.Printf("[webhook_update_checkout_details] missing consentId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "consentId is required"})
		return
	}
	if strings.TrimSpace(req.Issuer) == "" {
		log.Printf("[webhook_update_checkout_details] missing issuer")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "issuer is required"})
		return
	}

	// Acquirer business logic: e.g. persist checkoutId <-> consentEndToEndId/consentId mapping for redirect reconciliation.
	// Replace with your storage or downstream processing.
	_ = req

	log.Printf("[webhook_update_checkout_details] Checkout details received: checkoutId=%s consentEndToEndId=%s consentId=%s issuer=%s",
		req.CheckoutId, req.ConsentEndToEndId, req.ConsentId, req.Issuer)

	// Acquirer shall provide an acknowledgement back to API Gateway (per API spec).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "OK"})
}
