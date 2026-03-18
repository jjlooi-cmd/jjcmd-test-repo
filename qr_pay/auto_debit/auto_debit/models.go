package auto_debit

// AutoDebitRequest is the request payload for POST /v1/bw/autodebit (Initiate DuitNow AutoDebit).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/duitnow-autodebit
type AutoDebitRequest struct {
	// CheckoutID is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a payment intent.
	CheckoutID string `json:"checkoutId"` // Required; max length 36
	// ConsentID is the consent authorized for AutoDebit; from payment method details enquiry.
	ConsentID string `json:"consentId"` // Required; max length 35
	// Amount is the payment amount in two decimals, e.g. "10.00".
	Amount string `json:"amount"` // Required; max length 18
	// MerchantReferenceID is the payment reference to the recipient.
	MerchantReferenceID string `json:"merchantReferenceId"` // Required; max length 140
}

// AutoDebitResponse is the response body from POST /v1/bw/autodebit.
type AutoDebitResponse struct {
	Data    AutoDebitData `json:"data"`
	Message string        `json:"message"` // Reason code or error message; see PayNet reason codes
}

// AutoDebitData holds messageId and issuer from RPP.
type AutoDebitData struct {
	// MessageID is unique message identification from RPP; used to reconcile with RPP BackOffice or Reports.
	MessageID string `json:"messageId"` // max length 35
	// Issuer is the name of payer's issuing bank or wallet.
	Issuer string `json:"issuer"` // max length 100
}
