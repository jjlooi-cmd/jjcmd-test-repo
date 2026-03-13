package main

import (
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

	http.HandleFunc("/webhook/v2/account-lookup", printRequest)
	http.HandleFunc("/webhooks/v3/accounts/enquire-xc", printRequest)
	http.HandleFunc("/webhook/v3/account-lookup", printRequest)
	http.HandleFunc("/webhooks/v3/admin/event", printRequest)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
