package httpx

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Data  any            `json:"data,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
	Error *ErrorBody     `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON encodes data into an Envelope and writes it with the given HTTP status.
// Encoding happens before WriteHeader so that a marshaling failure can still
// return a 500 instead of a partial response.
func JSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(Envelope{Data: data})
	if err != nil {
		http.Error(w, "response encoding failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(body); err != nil {
		return // client disconnected after headers were already sent
	}
}

// Error writes a structured JSON error response.
func Error(w http.ResponseWriter, status int, code string, message string) {
	body, err := json.Marshal(Envelope{Error: &ErrorBody{Code: code, Message: message}})
	if err != nil {
		http.Error(w, "response encoding failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(body); err != nil {
		return // client disconnected after headers were already sent
	}
}
