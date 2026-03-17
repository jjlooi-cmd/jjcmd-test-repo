package webhook_update_checkout_details

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Handler implements POST /pg-router/rpp/v1/bw/notification/details for the PayNet Webhook: Update Checkout Details.
// This webhook maps the endToEndId to the checkoutId, allowing the acquirer to relate the endToEndId
// in the redirect URL back to the checkoutId when the issuer redirects with only the endToEndId (Step 19).
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/payment-intent#webhook-update-checkout-details
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "POST required"})
		return
	}

	var req UpdateCheckoutDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[webhook_update_checkout_details] invalid JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Request body must be valid JSON"})
		return
	}
	defer r.Body.Close()

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
	if strings.TrimSpace(req.RtpEndToEndId) == "" {
		log.Printf("[webhook_update_checkout_details] missing rtpEndToEndId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "rtpEndToEndId is required"})
		return
	}
	if strings.TrimSpace(req.Issuer) == "" {
		log.Printf("[webhook_update_checkout_details] missing issuer")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "issuer is required"})
		return
	}
	if strings.TrimSpace(req.PaymentMethod) == "" {
		log.Printf("[webhook_update_checkout_details] missing paymentMethod")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "paymentMethod is required"})
		return
	}

	// Acquirer business logic: e.g. persist checkoutId -> rtpEndToEndId mapping for redirect reconciliation.
	// Replace with your storage or downstream processing.
	_ = req

	log.Printf("[webhook_update_checkout_details] Checkout details received: checkoutId=%s rtpEndToEndId=%s issuer=%s paymentMethod=%s",
		req.CheckoutId, req.RtpEndToEndId, req.Issuer, req.PaymentMethod)

	// Acquirer shall provide an acknowledgement back to API Gateway (per API spec).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "OK"})
}
