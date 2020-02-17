package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/starkandwayne/signalfire/config"
)

type Authorizer interface {
	Auth(http.Handler) http.Handler
	TypeName() string
}

func NewAuthorizer(conf config.Auth) (auth Authorizer, err error) {
	switch strings.ToLower(conf.Type) {
	case "none", "noop":
		auth = &NoopAuthorizer{}
	case "basic":
		if conf.Realm == "" {
			conf.Realm = "SignalFire"
		}
		auth = &BasicAuthorizer{
			Username: conf.Username,
			Password: conf.Password,
			Realm:    conf.Realm,
		}
	default:
		err = fmt.Errorf("Unknown auth type: %s", conf.Type)
	}

	return
}

type NoopAuthorizer struct{}

func (*NoopAuthorizer) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) },
	)
}

func (*NoopAuthorizer) TypeName() string { return "none" }

type BasicAuthorizer struct {
	Username string
	Password string
	Realm    string
}

func (b *BasicAuthorizer) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username, password, isBasicAuth := r.BasicAuth()
			if !isBasicAuth {
				w.Header().Add("WWW-Authenticate", b.Realm)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if username != b.Username || password != b.Password {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h.ServeHTTP(w, r)
		},
	)
}

func (*BasicAuthorizer) TypeName() string { return "basic" }
