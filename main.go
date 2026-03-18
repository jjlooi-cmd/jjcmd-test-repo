package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"example.com/sample-repo/network_admin/echo"
	"example.com/sample-repo/network_admin/event_notification"
	"example.com/sample-repo/network_admin/sign_off"
	"example.com/sample-repo/network_admin/sign_on"
	"example.com/sample-repo/qr_acquirer/account_enquire_xc"
	"example.com/sample-repo/qr_acquirer/payments_transfer_xc"
	issuer_enquire "example.com/sample-repo/qr_issuer/account_enquire_xc"
	issuer_enquire_trx "example.com/sample-repo/qr_issuer/enquire_trx"
	issuer_payments_reverse "example.com/sample-repo/qr_issuer/payments_reverse"
	issuer_transfer "example.com/sample-repo/qr_issuer/payments_transfer_xc"
	auto_debit "example.com/sample-repo/qr_pay/auto_debit/auto_debit"
	enquire_checkout "example.com/sample-repo/qr_pay/one_time_payment/enquire_checkout"
	enquire_payment_status_v2 "example.com/sample-repo/qr_pay/one_time_payment/enquire_payment_status_v2"
	get_bank_list "example.com/sample-repo/qr_pay/one_time_payment/get_bank_list"
	initiate_checkout "example.com/sample-repo/qr_pay/one_time_payment/initiate_checkout"
	payment_intent "example.com/sample-repo/qr_pay/one_time_payment/payment_intent"
	retreive_checkout_payment_status "example.com/sample-repo/qr_pay/one_time_payment/retreive_checkout_payment_status"
	webhook_update_checkout_details "example.com/sample-repo/qr_pay/one_time_payment/webhook_update_checkout_details"
	webhook_update_payment_status "example.com/sample-repo/qr_pay/one_time_payment/webhook_update_payment_status"
	initiate_consent "example.com/sample-repo/qr_pay/saving_payment_and_consent/initiate_consent"
	payment_method_status "example.com/sample-repo/qr_pay/saving_payment_and_consent/payment_method_status"
	refund "example.com/sample-repo/qr_pay/refund/refund"
	enquire_refund "example.com/sample-repo/qr_pay/refund/enquire_refund"
	webhook_update_refund_status "example.com/sample-repo/qr_pay/refund/webhook_update_refund_status"
	save_payment_method "example.com/sample-repo/qr_pay/saving_payment_and_consent/save_payment_method"
	terminate_consent "example.com/sample-repo/qr_pay/saving_payment_and_consent/terminate_consent"
	consent_checkout_webhook "example.com/sample-repo/qr_pay/saving_payment_and_consent/webhook_update_checkout_details"
)

func printRequest(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	fmt.Println("--- Incoming request ---")
	fmt.Printf("Method: %s\n", r.Method)
	fmt.Printf("URL: %s\n", r.URL.String())
	fmt.Println("Headers:")
	for k, v := range r.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	fmt.Printf("Body: %s\n", string(body))
	fmt.Println("------------------------")

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/pg-router/v1/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hello world jj"))
	})

	// GET /v1/trigger-enquire-xc — calls PayNet Issuer EnquireXC with sample request (trigger from browser/curl).
	http.HandleFunc("/v1/trigger-enquire-xc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := issuer_enquire.DefaultClientConfig()
		req := issuer_enquire.SampleRequest()
		resp, statusCode, err := issuer_enquire.EnquireXC(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK) // still 200 so client gets body
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/trigger-transfer-xc — calls PayNet Issuer TransferXC (POST /v3/payments/transfer-xc) with sample request.
	http.HandleFunc("/v1/trigger-transfer-xc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := issuer_transfer.DefaultClientConfig()
		req := issuer_transfer.SampleRequest()
		resp, statusCode, respHeaders, err := issuer_transfer.TransferXC(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		// Print response headers and body to server log
		log.Printf("[trigger-transfer-xc] --- Response headers ---")
		for k, v := range respHeaders {
			log.Printf("[trigger-transfer-xc]   %s: %s", k, v)
		}
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		log.Printf("[trigger-transfer-xc] --- Response body ---\n%s", string(respJSON))
		// Include both in HTTP response
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":              true,
			"http_status":     statusCode,
			"response_header": respHeaders,
			"response":        resp,
		})
	})

	// GET /v1/trigger-enquire-trx — calls PayNet /v3/transactions/enquire with sample request.
	http.HandleFunc("/v1/trigger-enquire-trx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := issuer_enquire_trx.DefaultClientConfig()
		req := issuer_enquire_trx.SampleRequest()
		resp, statusCode, err := issuer_enquire_trx.EnquireTrx(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/trigger-sign-on — calls PayNet /v3/admin/sign-on (establish connectivity to RPP, txn code 000).
	http.HandleFunc("/v1/trigger-sign-on", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := sign_on.DefaultClientConfig()
		req := sign_on.SampleRequest()
		resp, statusCode, err := sign_on.SignOn(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/trigger-sign-off — calls PayNet /v3/admin/sign-off (disconnect from RPP, txn code 000).
	http.HandleFunc("/v1/trigger-sign-off", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := sign_off.DefaultClientConfig()
		req := sign_off.SampleRequest()
		resp, statusCode, err := sign_off.SignOff(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/trigger-echo — calls PayNet /v3/admin/echo (System Administration).
	http.HandleFunc("/v1/trigger-echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := echo.DefaultClientConfig()
		req := echo.SampleRequest()
		resp, statusCode, err := echo.Echo(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/trigger-reverse — calls PayNet v3/payments/reverse (reversal) with sample request.
	http.HandleFunc("/v1/trigger-reverse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := issuer_payments_reverse.DefaultClientConfig()
		req := issuer_payments_reverse.SampleRequest()
		resp, statusCode, err := issuer_payments_reverse.Reverse(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-payment-intent — calls PayNet DuitNow Pay POST /v1/payment/intent with sample request.
	http.HandleFunc("/v1/duitnowpay/trigger-payment-intent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := payment_intent.DefaultClientConfig()
		req := payment_intent.SampleRequest()
		resp, statusCode, err := payment_intent.CreatePaymentIntent(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-initiate-consent — calls PayNet DuitNow Pay POST /v1/bw/consent (Initiate Consent - self-hosted save payment method).
	http.HandleFunc("/v1/duitnowpay/trigger-initiate-consent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := initiate_consent.DefaultClientConfig()
		req := initiate_consent.SampleRequest()
		resp, statusCode, err := initiate_consent.InitiateConsent(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-payment-method-status — PayNet DuitNow Pay: Enquire Payment Method Status or Enquire Payment Method Details.
	// Query params: checkoutId → GET /v1/bw/consent/request?checkoutId=... (status). consentId → GET /v1/bw/consent?consentId=... (details).
	// If consentId is provided it takes precedence; otherwise checkoutId is used (defaults to sample). Status API rate limit: once every 30s per transaction.
	http.HandleFunc("/v1/duitnowpay/trigger-payment-method-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := payment_method_status.DefaultClientConfig()
		w.Header().Set("Content-Type", "application/json")

		consentId := r.URL.Query().Get("consentId")
		if consentId != "" {
			resp, statusCode, err := payment_method_status.GetPaymentMethodDetails(cfg, consentId)
			if err != nil {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"ok":          false,
					"error":       err.Error(),
					"http_status": statusCode,
				})
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          true,
				"http_status": statusCode,
				"response":    resp,
			})
			return
		}

		checkoutId := r.URL.Query().Get("checkoutId")
		if checkoutId == "" {
			checkoutId = payment_method_status.SampleCheckoutId()
		}
		resp, statusCode, err := payment_method_status.GetPaymentMethodStatus(cfg, checkoutId)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-save-payment-method — calls PayNet DuitNow Pay POST /v1/payment/intent with dataType "02" (Save Payment Method - DuitNow Consent).
	http.HandleFunc("/v1/duitnowpay/trigger-save-payment-method", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := save_payment_method.DefaultClientConfig()
		req := save_payment_method.SampleRequest()
		resp, statusCode, err := save_payment_method.CreateSavePaymentMethod(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-refund — calls PayNet DuitNow Pay POST /v1/bw/refund (Initiate Payment Refund).
	http.HandleFunc("/v1/duitnowpay/trigger-refund", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := refund.DefaultClientConfig()
		req := refund.SampleRequest()
		resp, statusCode, err := refund.InitiateRefund(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-enquire-refund — calls PayNet DuitNow Pay GET /v1/bw/refund?refundId=... (Enquire Refund Status).
	// Optional query param: refundId (defaults to sample value from API spec). Perform at least one hour after initial refund request.
	http.HandleFunc("/v1/duitnowpay/trigger-enquire-refund", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		refundId := r.URL.Query().Get("refundId")
		if refundId == "" {
			refundId = enquire_refund.SampleRefundId()
		}
		cfg := enquire_refund.DefaultClientConfig()
		resp, statusCode, err := enquire_refund.EnquireRefundStatus(cfg, refundId)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-autodebit — calls PayNet DuitNow Pay POST /v1/bw/autodebit (Initiate DuitNow AutoDebit).
	// Optional query params: checkoutId, consentId, amount (default to sample values from API spec when omitted).
	http.HandleFunc("/v1/duitnowpay/trigger-autodebit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		q := r.URL.Query()
		req := auto_debit.SampleRequest()
		if v := q.Get("checkoutId"); v != "" {
			req.CheckoutID = v
		}
		if v := q.Get("consentId"); v != "" {
			req.ConsentID = v
		}
		if v := q.Get("amount"); v != "" {
			req.Amount = v
		}
		cfg := auto_debit.DefaultClientConfig()
		resp, statusCode, err := auto_debit.InitiateAutoDebit(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-terminate-consent — calls PayNet DuitNow Pay DELETE /v1/bw/consent?consentId=... (Terminate Consent).
	// Optional query param: consentId (defaults to sample value from API spec).
	http.HandleFunc("/v1/duitnowpay/trigger-terminate-consent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		consentId := r.URL.Query().Get("consentId")
		if consentId == "" {
			consentId = terminate_consent.SampleConsentId()
		}
		cfg := terminate_consent.DefaultClientConfig()
		resp, statusCode, err := terminate_consent.TerminateConsent(cfg, consentId)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-initiate-checkout — calls PayNet DuitNow Pay POST /v1/bw/checkout (self-hosted page initiate checkout).
	http.HandleFunc("/v1/duitnowpay/trigger-initiate-checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := initiate_checkout.DefaultClientConfig()
		req := initiate_checkout.SampleRequest()
		resp, statusCode, err := initiate_checkout.InitiateCheckout(cfg, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-get-bank-list — calls PayNet DuitNow Pay GET /v2/bw/banks (bank list).
	http.HandleFunc("/v1/duitnowpay/trigger-get-bank-list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		cfg := get_bank_list.DefaultClientConfig()
		resp, statusCode, err := get_bank_list.GetBankList(cfg)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-enquire-checkout — calls PayNet DuitNow Pay GET /v1/bw/checkout?endToEndId=... (Enquire Checkout Details).
	// Optional query param: endToEndId (defaults to sample value from API spec).
	http.HandleFunc("/v1/duitnowpay/trigger-enquire-checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		endToEndId := r.URL.Query().Get("endToEndId")
		if endToEndId == "" {
			endToEndId = enquire_checkout.SampleEndToEndId()
		}
		cfg := enquire_checkout.DefaultClientConfig()
		resp, statusCode, err := enquire_checkout.EnquireCheckout(cfg, endToEndId)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-retrieve-checkout-payment-status — calls PayNet DuitNow Pay GET /v1/bw/rtp?checkoutId=... (Enquire Payment Status).
	// Optional query param: checkoutId (defaults to sample value from API spec). Rate limit: once every 30 seconds per transaction.
	http.HandleFunc("/v1/duitnowpay/trigger-retrieve-checkout-payment-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		checkoutId := r.URL.Query().Get("checkoutId")
		if checkoutId == "" {
			checkoutId = retreive_checkout_payment_status.SampleCheckoutId()
		}
		cfg := retreive_checkout_payment_status.DefaultClientConfig()
		resp, statusCode, err := retreive_checkout_payment_status.RetrievePaymentStatus(cfg, checkoutId)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	// GET /v1/duitnowpay/trigger-enquire-payment-status-v2 — calls PayNet DuitNow Pay GET /v2/bw/checkout-status?paymentMethod=01&checkoutId=... (Enquire Payment Status v2).
	// Optional query params: paymentMethod (default "01"), checkoutId (defaults to sample value). Rate limit: once every 30 seconds per transaction.
	http.HandleFunc("/v1/duitnowpay/trigger-enquire-payment-status-v2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "GET required"})
			return
		}
		paymentMethod := r.URL.Query().Get("paymentMethod")
		if paymentMethod == "" {
			paymentMethod = enquire_payment_status_v2.PaymentMethodOneTime
		}
		checkoutId := r.URL.Query().Get("checkoutId")
		if checkoutId == "" {
			checkoutId = enquire_payment_status_v2.SampleCheckoutId()
		}
		cfg := enquire_payment_status_v2.DefaultClientConfig()
		resp, statusCode, err := enquire_payment_status_v2.EnquirePaymentStatusV2(cfg, paymentMethod, checkoutId)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":          false,
				"error":       err.Error(),
				"http_status": statusCode,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"http_status": statusCode,
			"response":    resp,
		})
	})

	http.HandleFunc("/pg-router/webhooks/v2/account-lookup", printRequest)
	http.HandleFunc("/pg-router/webhooks/v3/accounts/enquire-xc", account_enquire_xc.Handler)
	http.HandleFunc("/pg-router/webhooks/v3/payments/transfer-xc", payments_transfer_xc.Handler)
	http.HandleFunc("/pg-router/webhooks/v3/account-lookup", printRequest)
	http.HandleFunc("/pg-router/webhooks/v3/admin/event", event_notification.Handler)
	http.HandleFunc("/webhooks/v3/admin/event", event_notification.Handler)

	// DuitNow Pay Webhook: Update Checkout Details — maps endToEndId to checkoutId for redirect reconciliation.
	http.HandleFunc("/pg-router/v1/payments/callback/RPP/MY/Notification/PaymentStatus/bw/notification/rtp-ct", webhook_update_checkout_details.Handler)
	http.HandleFunc("/pg-router/v1/payments/redirect/obw/RPP/MY/Redirect/RTP/reject", webhook_update_payment_status.Handler)
	http.HandleFunc("/pg-router/v1/payments/redirect/obw/RPP/MY/Redirect/RTP/success", webhook_update_payment_status.Handler)

	// DuitNow Pay Webhook: Update Checkout Details (consent flow) — maps consentEndToEndId to checkoutId for redirect reconciliation.
	http.HandleFunc("/pg-router/v1/duitnowpay/consent-notification", consent_checkout_webhook.Handler)
	// DuitNow Pay Webhook: Update Refund Status — notifies acquirer with final refund status once processing is complete.
	http.HandleFunc("/pg-router/v1/duitnowpay/refund", webhook_update_refund_status.Handler)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
