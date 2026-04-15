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
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta,omitempty"`
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
	ErrorWithMeta(w, status, code, message, nil)
}

func ErrorWithMeta(w http.ResponseWriter, status int, code string, message string, meta map[string]any) {
	body, err := json.Marshal(Envelope{Error: &ErrorBody{Code: code, Message: message, Meta: meta}})
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
