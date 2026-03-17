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
	get_bank_list "example.com/sample-repo/qr_pay/one_time_payment/get_bank_list"
	initiate_checkout "example.com/sample-repo/qr_pay/one_time_payment/initiate_checkout"
	payment_intent "example.com/sample-repo/qr_pay/one_time_payment/payment_intent"
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

	http.HandleFunc("/pg-router/webhooks/v2/account-lookup", printRequest)
	http.HandleFunc("/pg-router/webhooks/v3/accounts/enquire-xc", account_enquire_xc.Handler)
	http.HandleFunc("/pg-router/webhooks/v3/payments/transfer-xc", payments_transfer_xc.Handler)
	http.HandleFunc("/pg-router/webhooks/v3/account-lookup", printRequest)
	http.HandleFunc("/pg-router/webhooks/v3/admin/event", event_notification.Handler)
	http.HandleFunc("/webhooks/v3/admin/event", event_notification.Handler)
	http.HandleFunc("/pg-router/v1/payments/callback/RPP/MY/Notification/PaymentStatus/bw/notification/rtp-ct", printRequest)
	http.HandleFunc("/pg-router/v1/payments/redirect/obw/RPP/MY/Redirect/RTP", printRequest)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
