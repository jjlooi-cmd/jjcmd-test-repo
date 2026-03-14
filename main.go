package main

import (
	"encoding/json"
	"example.com/sample-repo/qr_acquirer/account_enquire_xc"
	"example.com/sample-repo/qr_acquirer/payments_transfer_xc"
	issuer_enquire "example.com/sample-repo/qr_issuer/account_enquire_xc"
	issuer_payments_reverse "example.com/sample-repo/qr_issuer/payments_reverse"
	issuer_transfer "example.com/sample-repo/qr_issuer/payments_transfer_xc"
	"fmt"
	"io"
	"log"
	"net/http"
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
	http.HandleFunc("/v1/hello", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/webhook/v2/account-lookup", printRequest)
	http.HandleFunc("/webhooks/v3/accounts/enquire-xc", account_enquire_xc.Handler)
	http.HandleFunc("/webhooks/v3/payments/transfer-xc", payments_transfer_xc.Handler)
	http.HandleFunc("/webhook/v3/account-lookup", printRequest)
	http.HandleFunc("/webhooks/v3/admin/event", printRequest)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
