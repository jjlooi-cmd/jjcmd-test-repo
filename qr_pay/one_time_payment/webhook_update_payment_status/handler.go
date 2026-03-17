package webhook_update_payment_status

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler implements POST /pg-router/v1/payments/redirect/obw/RPP/MY/Redirect/RTP/reject for the PayNet Webhook: Update Payment Status.
// This webhook updates the acquirer on the status and details of a transaction (success or rejected).
// Acquirer shall provide an acknowledgement back to API Gateway.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/payment-intent#webhook-update-payment-status
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[webhook_update_payment_status] Webhook triggered: %s %s", r.Method, r.URL.String())
	if len(r.Header) > 0 {
		log.Printf("[webhook_update_payment_status] Headers:")
		for k, v := range r.Header {
			log.Printf("[webhook_update_payment_status]   %s: %s", k, strings.Join(v, ", "))
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[webhook_update_payment_status] failed to read body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read request body"})
		return
	}
	defer r.Body.Close()
	if len(body) > 0 {
		log.Printf("[webhook_update_payment_status] Body (raw): %s", string(body))
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "POST required"})
		return
	}

	var req UpdatePaymentStatusRequest
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
		log.Printf("[webhook_update_payment_status] invalid JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Request body must be valid JSON"})
		return
	}

	log.Printf("[webhook_update_payment_status] --- Incoming request ---")
	log.Printf("[webhook_update_payment_status] Method: %s URL: %s", r.Method, r.URL.String())
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[webhook_update_payment_status] Body:\n%s", string(bodyBytes))
	log.Printf("[webhook_update_payment_status] ------------------------")

	if strings.TrimSpace(req.CheckoutId) == "" {
		log.Printf("[webhook_update_payment_status] missing checkoutId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "checkoutId is required"})
		return
	}
	if strings.TrimSpace(req.EndToEndId) == "" {
		log.Printf("[webhook_update_payment_status] missing endToEndId")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "endToEndId is required"})
		return
	}
	if strings.TrimSpace(req.PaymentStatus.Code) == "" {
		log.Printf("[webhook_update_payment_status] missing paymentStatus.code")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "paymentStatus.code is required"})
		return
	}
	if strings.TrimSpace(req.PaymentStatus.Substate) == "" {
		log.Printf("[webhook_update_payment_status] missing paymentStatus.substate")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "paymentStatus.substate is required"})
		return
	}
	if strings.TrimSpace(req.PaymentStatus.Message) == "" {
		log.Printf("[webhook_update_payment_status] missing paymentStatus.message")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "paymentStatus.message is required"})
		return
	}
	if strings.TrimSpace(req.Issuer) == "" {
		log.Printf("[webhook_update_payment_status] missing issuer")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "issuer is required"})
		return
	}
	if strings.TrimSpace(req.PaymentMethod) == "" {
		log.Printf("[webhook_update_payment_status] missing paymentMethod")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "paymentMethod is required"})
		return
	}

	// Acquirer business logic: e.g. persist payment status, update order state, notify user.
	// Replace with your storage or downstream processing.
	_ = req

	log.Printf("[webhook_update_payment_status] Payment status received: checkoutId=%s endToEndId=%s code=%s substate=%s issuer=%s paymentMethod=%s",
		req.CheckoutId, req.EndToEndId, req.PaymentStatus.Code, req.PaymentStatus.Substate, req.Issuer, req.PaymentMethod)

	// Acquirer shall provide an acknowledgement back to API Gateway (per API spec).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "OK"})
}
