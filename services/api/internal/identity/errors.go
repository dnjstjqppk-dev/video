package identity

import (
	"encoding/json"
	"net/http"
)

type ApiErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Field     string `json:"field,omitempty"`
}

type ApiErrorResponse struct {
	Error ApiErrorDetail `json:"error"`
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = "req_generated"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ApiErrorResponse{
		Error: ApiErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: reqID,
		},
	})
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
