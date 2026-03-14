package event_notification

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/sample-repo/jws_generation"
)

// JWS issuer and credential key for response signing (participant BIC / onboarding key).
const (
	jwsIssuer        = "MBBEMYKL"
	jwsCredentialKey = "64feb830"
	// participantBic is used as BIC in response businessMessageId (participant = API consumer).
	participantBic = "MBBEMYKL"
)

// Business message ID format: YYYYMMDD(8) + BIC(8) + TxnCode(3) + Originator(1) + Channel(2) + Sequence(8).
const (
	bicStartPosition       = 8
	bicLength              = 8
	originatorCodePosition = 19
)

// Handler implements POST /webhooks/v3/admin/event for PayNet System Event Notification (Network Administration - Acquirer).
// Ref: document (6).yaml — webhooks /webhooks/v3/admin/event
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[event_notification] Incoming request Authorization (token): %s", r.Header.Get("Authorization"))
	if r.Method != http.MethodPost {
		setEventResponseHeaders(w, r, "")
		writeEventJSON(w, http.StatusMethodNotAllowed, EventWebhookResponse{})
		return
	}

	var req EventWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[event_notification] invalid JSON: %v", err)
		setEventResponseHeaders(w, r, "")
		writeEventJSON(w, http.StatusBadRequest, EventWebhookResponse{})
		return
	}
	defer r.Body.Close()

	log.Printf("[event_notification] --- Incoming request ---")
	log.Printf("[event_notification] Method: %s URL: %s", r.Method, r.URL.String())
	log.Printf("[event_notification] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		r.Header.Get("X-Client-Id"),
		r.Header.Get("X-Api-Version"),
		r.Header.Get("x-business-message-id"))
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[event_notification] Body:\n%s", string(bodyBytes))
	log.Printf("[event_notification] ------------------------")

	// Required field: eventCode
	if strings.TrimSpace(req.EventCode) == "" {
		log.Printf("[event_notification] missing eventCode")
		setEventResponseHeaders(w, r, "")
		writeEventJSON(w, http.StatusBadRequest, EventWebhookResponse{})
		return
	}

	originalBizMsgId := ""
	if req.AppHeader != nil {
		originalBizMsgId = strings.TrimSpace(req.AppHeader.BusinessMessageId)
	}
	if originalBizMsgId == "" {
		originalBizMsgId = r.Header.Get("x-business-message-id")
	}
	responseBizMsgId := responseBusinessMessageId(originalBizMsgId, participantBic)
	creationDateTime := malaysiaTimeNow()

	resp := EventWebhookResponse{
		AppHeader: EventWebhookResponseAppHeader{
			BusinessMessageId:         responseBizMsgId,
			CreationDateTime:          creationDateTime,
			OriginalBusinessMessageId: originalBizMsgId,
		},
		EventCode: req.EventCode,
	}

	bodyRespBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[event_notification] marshal response: %v", err)
		setEventResponseHeaders(w, r, responseBizMsgId)
		writeEventJSON(w, http.StatusOK, resp)
		return
	}

	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Printf("[event_notification] load private key for JWS: %v", err)
		setEventResponseHeaders(w, r, responseBizMsgId)
		writeEventJSON(w, http.StatusOK, resp)
		return
	}
	token, err := jws_generation.GenerateJWS(jws_generation.GenerateOptions{
		PrivateKey:        privateKey,
		Algorithm:         jws_generation.RS512,
		Issuer:            jwsIssuer,
		BusinessMessageID: responseBizMsgId,
		CredentialKey:     jwsCredentialKey,
		PayloadForHash:    bodyRespBytes,
	})
	if err != nil {
		log.Printf("[event_notification] generate JWS: %v", err)
		setEventResponseHeaders(w, r, responseBizMsgId)
		writeEventJSON(w, http.StatusOK, resp)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	setEventResponseHeaders(w, r, responseBizMsgId)
	writeEventResponse(w, http.StatusOK, bodyRespBytes)
}

func responseBusinessMessageId(requestId string, participantBic string) string {
	if requestId == "" {
		return ""
	}
	if len(requestId) <= originatorCodePosition {
		return requestId
	}
	b := []byte(requestId)
	participantBic = strings.TrimSpace(participantBic)
	if participantBic != "" && len(b) >= bicStartPosition+bicLength {
		bic := participantBic
		if len(bic) > bicLength {
			bic = bic[:bicLength]
		} else if len(bic) < bicLength {
			bic = bic + strings.Repeat(" ", bicLength-len(bic))
		}
		copy(b[bicStartPosition:bicStartPosition+bicLength], bic)
	}
	b[originatorCodePosition] = 'R'
	return string(b)
}

func malaysiaTimeNow() string {
	// Malaysian time = UTC+8
	return time.Now().UTC().Add(8 * time.Hour).Format("2006-01-02T15:04:05.000+08:00")
}

func setEventResponseHeaders(w http.ResponseWriter, r *http.Request, businessMessageId string) {
	w.Header().Set("Content-Type", "application/json")
	if v := r.Header.Get("X-Client-Id"); v != "" {
		w.Header().Set("X-Client-Id", v)
	} else if businessMessageId != "" {
		w.Header().Set("X-Client-Id", participantBic)
	}
	if v := r.Header.Get("X-Api-Version"); v != "" {
		w.Header().Set("X-Api-Version", v)
	}
	if businessMessageId != "" {
		w.Header().Set("x-business-message-id", businessMessageId)
	}
}

func writeEventJSON(w http.ResponseWriter, statusCode int, body EventWebhookResponse) {
	bodyBytes, _ := json.Marshal(body)
	log.Printf("[event_notification] --- Outgoing response ---")
	log.Printf("[event_notification] HTTP %d", statusCode)
	log.Printf("[event_notification] Body:\n%s", string(bodyBytes))
	log.Printf("[event_notification] -------------------------")
	w.WriteHeader(statusCode)
	_, _ = w.Write(bodyBytes)
}

// writeEventResponse writes the exact response bytes (used when body was already marshaled for JWS).
func writeEventResponse(w http.ResponseWriter, statusCode int, bodyBytes []byte) {
	log.Printf("[event_notification] --- Outgoing response ---")
	log.Printf("[event_notification] HTTP %d", statusCode)
	log.Printf("[event_notification] Body:\n%s", string(bodyBytes))
	log.Printf("[event_notification] -------------------------")
	w.WriteHeader(statusCode)
	_, _ = w.Write(bodyBytes)
}
