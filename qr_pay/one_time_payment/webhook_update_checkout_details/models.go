package webhook_update_checkout_details

// UpdateCheckoutDetailsRequest is the request body for the PayNet Webhook: Update Checkout Details.
// This webhook maps the endToEndId to the checkoutId so the acquirer can relate the endToEndId
// in the redirect URL back to the checkoutId when the issuer redirects with only the endToEndId.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/payment-intent#webhook-update-checkout-details
type UpdateCheckoutDetailsRequest struct {
	// CheckoutId is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a payment intent. Max length: 36.
	CheckoutId string `json:"checkoutId"`
	// RtpEndToEndId is the unique message identification from RPP. Can be used to reconcile with RPP BackOffice or Reports. Max length: 35.
	RtpEndToEndId string `json:"rtpEndToEndId"`
	// Issuer is the name of payer's issuing bank / wallet. Max length: 100.
	Issuer string `json:"issuer"`
	// PaymentMethod is the payer selected payment method. 01 = DuitNow Online Banking / Wallets. Max length: 2.
	PaymentMethod string `json:"paymentMethod"`
}
