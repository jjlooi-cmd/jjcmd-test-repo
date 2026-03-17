package get_bank_list

// GetBankListResponse is the response body from GET /v2/bw/banks (DuitNow Pay Get Bank List).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/get-bank-list
// Note: API returns data as an array of objects (each with retail and corporate lists).
type GetBankListResponse struct {
	Data    []GetBankListData `json:"data"`    // Array of { retail, corporate } per segment
	Message string            `json:"message"` // Reason code, e.g. "U000"
}

// GetBankListData holds retail and corporate bank lists.
type GetBankListData struct {
	Retail    []BankEntry `json:"retail"`
	Corporate []BankEntry `json:"corporate"`
}

// BankEntry represents one issuing participant (retail or corporate) in the bank list.
// Fields match PayNet Get Bank List response; corporate entries may have null isFpx, isBank.
type BankEntry struct {
	BICCode      string  `json:"bicCode"`      // Bank Identification Code, max 35
	Name         string  `json:"name"`         // Name of issuing participant, max 35
	Browser      string  `json:"browser"`       // Browser info/URL, max 140
	AndroidAppID string  `json:"androidAppId"` // Android app identifier, max 140
	IOSAppID     string  `json:"iosAppId"`     // iOS app identifier, max 140
	URL          string  `json:"url"`          // Bank URL for redirect, max 140
	IsConsent    bool    `json:"isConsent"`    // Supports DuitNow Consent
	IsActive     bool    `json:"isActive"`     // Active in production
	IsFpx        *bool   `json:"isFpx"`        // Supports FPX; may be null for corporate
	IsBank       *bool   `json:"isBank"`       // true=bank, false=e-wallet; may be null for corporate
	IsObw        bool    `json:"isObw"`        // Supports DuitNow Online Banking/Wallets
	Priority     int     `json:"priority"`     // Sort order; lower = higher priority; 0 = sort alphabetically
}
