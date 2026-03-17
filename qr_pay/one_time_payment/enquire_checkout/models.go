package enquire_checkout

// EnquireCheckoutResponse is the response body from GET /v1/bw/checkout?endToEndId=... (Enquire Checkout Details).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/checkout-details
// The "data" object varies: one-time payment has checkoutId, rtpEndToEndId, issuer, paymentMethod;
// save payment method has checkoutId, consentEndToEndId, consentId, issuer.
type EnquireCheckoutResponse struct {
	Data    EnquireCheckoutData `json:"data"`
	Message string              `json:"message"` // "OK" if successful; otherwise reason code
}

// EnquireCheckoutData holds checkout details. For one-time payment: RtpEndToEndId and PaymentMethod are set.
// For save payment method: ConsentEndToEndId and ConsentId are set. CheckoutId and Issuer are always present.
type EnquireCheckoutData struct {
	// CheckoutId is the unique external identifier (uuid v4) provided by the acquirer to PayNet. Max length 36.
	CheckoutId string `json:"checkoutId"`
	// RtpEndToEndId is set for one-time payment: unique message identification from RPP. Max length 35.
	RtpEndToEndId string `json:"rtpEndToEndId,omitempty"`
	// ConsentEndToEndId is set for save payment method: unique message identification from RPP. Max length 35.
	ConsentEndToEndId string `json:"consentEndToEndId,omitempty"`
	// ConsentId is set for save payment method: consent authorized for AutoDebit. Max length 36.
	ConsentId string `json:"consentId,omitempty"`
	// Issuer is the name of payer's issuing bank/wallet. Max length 100.
	Issuer string `json:"issuer"`
	// PaymentMethod is set for one-time payment: "01" = DuitNow Online Banking/Wallets. Max length 2.
	PaymentMethod string `json:"paymentMethod,omitempty"`
}
