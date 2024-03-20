package collect

import (
	"encoding/json"
	"net/http"
)

type errHTTP struct {
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"error"`
}

func (e errHTTP) Error() string {
	return e.Message
}

func (e errHTTP) JSON() string {
	b, _ := json.Marshal(&e)
	return string(b)
}

var errHTTPInternalError = &errHTTP{http.StatusInternalServerError, "internal server error"}
