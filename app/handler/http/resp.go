package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// HTTPErrCode as returned by the server.
type HTTPErrCode string

// Known HTTPErrCode values.
const (
	HTTPErrCodeInternalError HTTPErrCode = "internal_server_error"
	HTTPErrCodeUnauthorized  HTTPErrCode = "unauthorized"
	HTTPErrCodeGone          HTTPErrCode = "gone"

	HTTPErrCodeUnfillableTable HTTPErrCode = "unfillable_table"
	HTTPErrCodeNoTables        HTTPErrCode = "no_tables"

	HTTPErrCodeNoAuthCode HTTPErrCode = "no_auth_code"

	HTTPErrCodeUnknownWorkspace HTTPErrCode = "unknown_workspace"
)

// HTTPErr returned by the server in case of errors.
type HTTPErr struct {
	Code        HTTPErrCode `json:"code"`
	Message     string      `json:"message"`
	Description string      `json:"description,omitempty"`
}

// NewHTTPErr constructor.
func NewHTTPErr(c HTTPErrCode, msg, desc string) HTTPErr {
	return HTTPErr{
		Code:        c,
		Message:     msg,
		Description: desc,
	}
}

// WriteHTTPErr to the response.
func WriteHTTPErr(w http.ResponseWriter, status int, he HTTPErr) {
	b, err := json.Marshal(&he)
	if err != nil {
		log.Printf("Handler: WriteHTTPErr: couldn't write response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b) // nolint: errcheck
}

// WriteInternalServerErr into the ResponseWriter.
func WriteInternalServerErr(w http.ResponseWriter) {
	WriteHTTPErr(w, http.StatusInternalServerError, NewHTTPErr(HTTPErrCodeInternalError, "", ""))
}
