package enquire_refund

// EnquireRefundStatusResponse is the response body from GET /v1/bw/refund?refundId=... (Enquire Refund Status).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/refund-status
type EnquireRefundStatusResponse struct {
	Data    EnquireRefundStatusData `json:"data"`
	Message string                 `json:"message"` // Reason code; see PayNet reason codes
}

// EnquireRefundStatusData holds the refund status data from PayNet.
// refundStatus: REFUND_PENDING, REFUND_ACCEPTED, REFUND_REJECTED, REFUND_EXCEPTION.
type EnquireRefundStatusData struct {
	CheckoutID   string `json:"checkoutId"`   // Same checkoutId from first payment; max length 36
	EndToEndID   string `json:"endToEndId"`  // Unique message identification from RPP; max length 35
	Code         string `json:"code"`         // Status code; see PayNet status codes; max length 4
	RefundStatus string `json:"refundStatus"` // REFUND_PENDING, REFUND_ACCEPTED, REFUND_REJECTED, REFUND_EXCEPTION; max length 35
	Issuer       string `json:"issuer"`       // Name of payer's issuing bank; max length 100
	PaymentMethod string `json:"paymentMethod"` // Payment method from first payment; max length 2
	Amount       string `json:"amount"`       // Amount requested for refund; max length 18
}
