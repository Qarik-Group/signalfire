package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type APIWebDev struct {
	filePath string
	mimeType string
}

func newAPIWebDev(filePath, mimeType string) *APIWebDev {
	return &APIWebDev{
		filePath: filePath,
		mimeType: mimeType,
	}
}

func (a *APIWebDev) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contents, err := ioutil.ReadFile(a.filePath)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError,
			APIError{Error: fmt.Sprintf("Could not read mapped file: %s", err)})
		return
	}
	w.Header().Add("Content-Type", a.mimeType)
	w.WriteHeader(http.StatusOK)
	w.Write(contents)
}
