package terminate_consent

// TerminateConsentResponse is the response body from DELETE /v1/bw/consent?consentId=... (Terminate Consent).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/terminate-consent
type TerminateConsentResponse struct {
	Data    TerminateConsentData `json:"data"`
	Message string               `json:"message"` // Reason code e.g. "U000"; refer to reason codes.
}

// TerminateConsentData holds the messageId from RPP after consent is removed.
type TerminateConsentData struct {
	MessageId string `json:"messageId"` // Unique message identification from RPP. Max length 35. Used for reconciliation.
}
