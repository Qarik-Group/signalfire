package server

import (
	"net/http"
)

type APIInfo struct {
	payload []byte
}

func NewAPIInfo(version, authType string) *APIInfo {
	j := jsonMustMarshal(struct {
		Version  string `json:"version"`
		AuthType string `json:"auth_type"`
	}{
		Version:  version,
		AuthType: authType,
	})
	return &APIInfo{
		payload: j,
	}
}

func (a *APIInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeResponseBytes(w, http.StatusOK, a.payload)
}
