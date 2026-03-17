package enquire_payment_status_v2

// EnquirePaymentStatusV2Response is the response body from GET /v2/bw/checkout-status (Enquire Payment Status, v2.0.0).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-status
// v2 optimises response by removing messageId from data.
type EnquirePaymentStatusV2Response struct {
	Data    EnquirePaymentStatusV2Data `json:"data"`
	Message string                     `json:"message"` // Reason code, refer to reason codes in appendix
}

// EnquirePaymentStatusV2Data holds the payment status data from PayNet v2.
// transactionStatus: RECEIVED (Pending), CLEARED (Successful Credit), REJECTED, PENDAUTH (Pending authorization).
// paymentMethod: 01 - DuitNow Online Banking / Wallets, 02 - Save payment method.
type EnquirePaymentStatusV2Data struct {
	EndToEndId         string `json:"endToEndId"`         // Unique message identification from RPP, max 35
	TransactionStatus string `json:"transactionStatus"`  // RECEIVED, CLEARED, REJECTED, PENDAUTH
	Issuer             string `json:"issuer"`            // Name of payer's issuing bank/wallet, max 100
	PaymentMethod      string `json:"paymentMethod"`     // 01 - DuitNow OB/Wallets, 02 - Save payment method, max 2
}
