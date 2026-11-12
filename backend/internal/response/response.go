package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	OK       bool   `json:"ok"`
	Message  string `json:"message,omitempty"`
	Filename string `json:"filename,omitempty"`
	URL      string `json:"url,omitempty"`
}

func JSON(w http.ResponseWriter, code int, body Body) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func OK(w http.ResponseWriter, body Body) {
	body.OK = true
	JSON(w, http.StatusOK, body)
}

func Error(w http.ResponseWriter, code int, msg string) {
	JSON(w, code, Body{OK: false, Message: msg})
}
