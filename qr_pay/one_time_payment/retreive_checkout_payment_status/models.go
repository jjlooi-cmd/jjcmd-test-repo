package retreive_checkout_payment_status

// RetrievePaymentStatusResponse is the response body from GET /v1/bw/rtp?checkoutId=... (Enquire Payment Status, v1.0.0).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-status
type RetrievePaymentStatusResponse struct {
	Data    RetrievePaymentStatusData `json:"data"`
	Message string                    `json:"message"` // Reason code, e.g. "U000"
}

// RetrievePaymentStatusData holds the payment status data from PayNet.
// transactionStatus: RECEIVED (Pending), CLEARED (Successful Credit), REJECTED, PENDAUTH (Pending authorization).
// paymentMethod: 01 - DuitNow Online Banking / Wallets.
type RetrievePaymentStatusData struct {
	MessageId          string `json:"messageId"`          // Unique message identification from RPP, max 35
	EndToEndId         string `json:"endToEndId"`         // Unique message identification from RPP, max 35
	TransactionStatus string `json:"transactionStatus"` // RECEIVED, CLEARED, REJECTED, PENDAUTH
	Issuer             string `json:"issuer"`             // Name of payer's issuing bank/wallet, max 100
	PaymentMethod      string `json:"paymentMethod"`      // 01 - DuitNow Online Banking / Wallets, max 2
}
