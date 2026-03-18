package refund

// RefundRequest is the request payload for POST /v1/bw/refund (Initiate Payment Refund).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/initiate-refund#request
type RefundRequest struct {
	// RefundID is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a refund request.
	RefundID string `json:"refundId"` // Required; max length 36
	// CheckoutID is the same checkoutId used during the first payment initiated.
	CheckoutID string `json:"checkoutId"` // Required; max length 36
	// PaymentMethod is the paymentMethod selected during the first payment initiated (e.g. "01" for CASA).
	PaymentMethod string `json:"paymentMethod"` // Required; max length 2
	// Amount is the refund amount with two decimal places (e.g. "10.00").
	Amount string `json:"amount"` // Required; max length 18
	// MerchantReferenceID is the refund reference shown to the user for the details of the refund.
	MerchantReferenceID string `json:"merchantReferenceId"` // Required; max length 140
}

// RefundResponse is the response body from POST /v1/bw/refund.
type RefundResponse struct {
	Data    RefundData `json:"data"`
	Message string     `json:"message"` // Reason code; see PayNet reason codes
}

// RefundData holds endToEndId and refundStatus from RPP.
type RefundData struct {
	// EndToEndID is unique message identification from RPP; used to reconcile with RPP BackOffice or Reports.
	EndToEndID string `json:"endToEndId"` // max length 35
	// RefundStatus is REFUND_PENDING while processing; final status is delivered via webhook.
	RefundStatus string `json:"refundStatus"` // max length 35
}
