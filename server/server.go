package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/starkandwayne/signalfire/config"
	"github.com/starkandwayne/signalfire/core"
	"github.com/starkandwayne/signalfire/version"
)

type Server struct {
	collator *core.Collator
	server   *http.Server
}

func New(conf config.Server, collator *core.Collator) (*Server, error) {
	ret := &Server{
		collator: collator,
	}

	auth, err := NewAuthorizer(conf.Auth)
	if err != nil {
		return nil, fmt.Errorf("Error initializing server auth: %s", err)
	}

	ret.server = &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", conf.Port),
		Handler:           ret.newRouter(auth, ret.collator),
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}

	shouldTLS, err := ret.shouldUseTLS(conf.TLS.Certificate, conf.TLS.PrivateKey)
	if err != nil {
		return nil, err
	}
	if shouldTLS {
		ret.server.TLSConfig, err = ret.newTLSConfig(conf)
		if err != nil {
			return nil, fmt.Errorf("Could not configure TLS: %s", err)
		}
	}

	return ret, nil
}

func (s *Server) shouldUseTLS(cert, key string) (should bool, err error) {
	if cert != "" && key != "" {
		should = true
	} else if cert != "" {
		err = fmt.Errorf("Certificate provided without private key")
	} else if key != "" {
		err = fmt.Errorf("Private key provided without certificate")
	}

	return
}

func (s *Server) newRouter(auth Authorizer, col *core.Collator) http.Handler {
	ret := mux.NewRouter()

	notFoundHandler := NewAPINotFound()
	ret.NotFoundHandler = notFoundHandler
	ret.MethodNotAllowedHandler = notFoundHandler

	ret.Handle("/v1/info", NewAPIInfo(version.Version, auth.TypeName()))
	ret.Handle("/v1/deployment-groups", auth.Auth(NewAPIGroups(col)))

	return ret
}

func (s *Server) newTLSConfig(conf config.Server) (*tls.Config, error) {
	certificate, err := tls.X509KeyPair([]byte(conf.TLS.Certificate), []byte(conf.TLS.PrivateKey))
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{certificate},
	}, nil
}

func (s *Server) Run() error {
	if s.server.TLSConfig == nil {
		return s.server.ListenAndServe()
	}
	return s.server.ListenAndServeTLS("", "")
}

type APIError struct {
	Error string `json:"error"`
}

var (
	//InternalServerErrorMessagePayload is a string to be written to the response
	// body in the case of an internal server error
	InternalServerErrorMessagePayload = jsonMustMarshal(&APIError{Error: "An internal server error occurred"})
)

func jsonMustMarshal(obj interface{}) []byte {
	b, err := json.Marshal(&obj)
	if err != nil {
		panic(fmt.Sprintf("Could not marshal into JSON as expected: %s", err))
	}

	return b
}

//JSON marshals the given object and sends it over the response object
// If an error occurs while marshalling, a 500 response is sent, and an error
// is returned from the function
func writeResponse(w http.ResponseWriter, code int, obj interface{}) error {
	out, err := json.Marshal(&obj)
	if err != nil {
		code, out = http.StatusInternalServerError, InternalServerErrorMessagePayload
	}

	writeResponseBytes(w, code, out)
	return err
}

func writeResponseBytes(w http.ResponseWriter, code int, body []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body)
}

type APINotFound struct {
	payload []byte
}

func NewAPINotFound() *APINotFound {
	return &APINotFound{payload: jsonMustMarshal(APIError{Error: "Not found"})}
}

func (a *APINotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeResponseBytes(w, http.StatusNotFound, a.payload)
}
