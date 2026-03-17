package webhook_update_payment_status

// UpdatePaymentStatusRequest is the request body for the PayNet Webhook: Update Payment Status.
// This webhook updates the acquirer on the status and details of a transaction (success or rejected).
// For corporate flows (transactionFlow = "02"), perform Enquire Payment Status on the 5th day to confirm final status.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/payment-intent#webhook-update-payment-status
type UpdatePaymentStatusRequest struct {
	// CheckoutId is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a payment intent. Max length: 36.
	CheckoutId string `json:"checkoutId"`
	// EndToEndId is the unique message identification from RPP. Can be used to reconcile with RPP BackOffice or Reports. Max length: 35.
	EndToEndId string `json:"endToEndId"`
	// PaymentStatus contains the payment status details.
	PaymentStatus PaymentStatus `json:"paymentStatus"`
	// Issuer is the name of payer's issuing bank / wallet. Max length: 100.
	Issuer string `json:"issuer"`
	// PaymentMethod is the payer selected payment method. 01 = DuitNow Online Banking / Wallets. Max length: 2.
	PaymentMethod string `json:"paymentMethod"`
}

// PaymentStatus holds status and substate from the debiting bank.
// Substates: RECEIVED (Pending), CLEARED (Successful Credit), REJECTED, PENDAUTH (Pending authorization).
type PaymentStatus struct {
	// PayerName is the name of payer from the debiting bank. Max length: 100. Present for success/status updates; may be omitted for rejected.
	PayerName string `json:"payerName,omitempty"`
	// Code refers to status codes (e.g. ACTC, ACSP). Max length: 4.
	Code string `json:"code"`
	// Substate: RECEIVED, CLEARED, REJECTED, PENDAUTH.
	Substate string `json:"substate"`
	// Message refers to reason codes (e.g. U002, U000). Max length: 1024.
	Message string `json:"message"`
}
