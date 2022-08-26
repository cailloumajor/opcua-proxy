package api

import (
	"encoding/json"
	"net/http"
)

// ReplyProblemDetails replies with problem details, as of [RFC 7807].
//
// [RFC 7807]: https://datatracker.ietf.org/doc/html/rfc7807
func ReplyProblemDetails(w http.ResponseWriter, problemType, title, detail string, status int) {
	d := struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail,omitempty"`
	}{
		Type:   "/problem/" + problemType,
		Title:  title,
		Status: status,
		Detail: detail,
	}

	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&d)
}
