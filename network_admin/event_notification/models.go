package event_notification

// EventWebhookRequest is the request body for POST /webhooks/v3/admin/event (System Event Notification).
// Ref: document (6).yaml — webhooks /webhooks/v3/admin/event
type EventWebhookRequest struct {
	AppHeader         *EventAppHeader `json:"appHeader,omitempty"`
	EventCode         string          `json:"eventCode"`
	EventParameter    []interface{}   `json:"eventParameter,omitempty"`
	EventDescription  string          `json:"eventDescription,omitempty"`
	EventTime         string          `json:"eventTime,omitempty"`
}

// EventAppHeader is the application header for event request/response.
type EventAppHeader struct {
	BusinessMessageId string `json:"businessMessageId"`
	CreationDateTime  string `json:"creationDateTime"`
}

// EventWebhookResponse is the response body for POST /webhooks/v3/admin/event (System Event Acknowledgement).
type EventWebhookResponse struct {
	AppHeader EventWebhookResponseAppHeader `json:"appHeader"`
	EventCode string                        `json:"eventCode"`
}

// EventWebhookResponseAppHeader is the application header in the acknowledgement response.
type EventWebhookResponseAppHeader struct {
	BusinessMessageId         string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
}
