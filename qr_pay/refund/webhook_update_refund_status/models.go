package webhook_update_refund_status

// UpdateRefundStatusRequest is the request body for the PayNet Webhook: Update Refund Status.
// This webhook notifies the acquirer with the final status of the refund once processing is complete.
// Acquirer shall provide an acknowledgement back to API Gateway.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/initiate-refund#webhook--update-refund-status
type UpdateRefundStatusRequest struct {
	// RefundId is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a refund request. Max length: 36.
	RefundId string `json:"refundId"`
	// CheckoutId is the same checkoutId used during the first payment initiated. Max length: 36.
	CheckoutId string `json:"checkoutId"`
	// EndToEndId is the unique message identification from RPP. Can be used to reconcile with RPP BackOffice or Reports. Max length: 35.
	EndToEndId string `json:"endToEndId"`
	// PaymentStatus contains the final refund status and reason code.
	PaymentStatus PaymentStatus `json:"paymentStatus"`
	// Issuer is the name of payer's issuing bank. Max length: 100.
	Issuer string `json:"issuer"`
	// PaymentMethod is the paymentMethod selected during the first payment initiated. Max length: 2.
	PaymentMethod string `json:"paymentMethod"`
	// Amount is the amount requested for the refund. Max length: 18.
	Amount string `json:"amount"`
}

// PaymentStatus holds the refund status from the webhook.
// Substate: REFUND_PENDING, REFUND_ACCEPTED, REFUND_REJECTED, REFUND_EXCEPTION.
type PaymentStatus struct {
	// Code refers to status codes (e.g. ACTC). Max length: 4.
	Code string `json:"code"`
	// Substate is the final refund status. Max length: 35.
	// REFUND_PENDING - Refund is in progress.
	// REFUND_ACCEPTED - Refund sent to issuing bank for processing.
	// REFUND_REJECTED - Refund unable to process.
	// REFUND_EXCEPTION - Unexpected error during refund; try again and inform PayNet if problem persists.
	Substate string `json:"substate"`
	// Message refers to reason codes. Max length: 1024.
	Message string `json:"message"`
}
