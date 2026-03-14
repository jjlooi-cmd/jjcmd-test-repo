package echo

// SystemAdminRequest is the request payload for POST /v3/admin/echo (Network Administration - Acquirer).
// Schema from PayNet Network Administration API.
// Ref: document (6).yaml — /v3/admin/echo
type SystemAdminRequest struct {
	AppHeader SystemAdminRequestAppHeader `json:"appHeader"`
}

// SystemAdminRequestAppHeader is the application header for echo request.
type SystemAdminRequestAppHeader struct {
	BusinessMessageId string `json:"businessMessageId"`
	CreationDateTime  string `json:"creationDateTime"`
}

// SystemAdminResponse is the response body from POST /v3/admin/echo.
type SystemAdminResponse struct {
	AppHeader SystemAdminResponseAppHeader `json:"appHeader"`
	Resp      SystemAdminResp               `json:"resp"`
}

// SystemAdminResponseAppHeader is the application header in the response.
type SystemAdminResponseAppHeader struct {
	BusinessMessageId         string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
}

// SystemAdminResp is the response status (ACTC = Accepted, RJCT = Rejected).
type SystemAdminResp struct {
	Status string `json:"status"` // ACTC | RJCT
}
