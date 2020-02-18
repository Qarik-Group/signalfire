package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/starkandwayne/signalfire/config"
	"github.com/starkandwayne/signalfire/log"
)

type Authorizer interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	TypeName() string
}

type AuthorizerComponents struct {
	log          *log.Logger
	tokenChecker *tokenChecker
}

func NewAuthorizer(conf config.Auth, t *tokenChecker) (auth Authorizer, err error) {
	switch strings.ToLower(conf.Type) {
	case "none", "noop":
		auth = newNoopAuthorizer(t)
	case "basic":
		if conf.Realm == "" {
			conf.Realm = "SignalFire"
		}
		auth = newBasicAuthorizer(t, basicAuthorizerConfig{
			Username: conf.Username,
			Password: conf.Password,
			Realm:    conf.Realm,
		})
	default:
		err = fmt.Errorf("Unknown auth type: %s", conf.Type)
	}

	return
}

func writeSessionResponse(w http.ResponseWriter, t *tokenChecker) {
	sessionToken, err := t.newSession()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, APIError{Error: err.Error()})
	}
	writeResponse(w, http.StatusOK, AuthTokenResponse{Token: sessionToken})
}

type AuthTokenResponse struct {
	Token string `json:"token"`
}

type NoopAuthorizer struct {
	t *tokenChecker
}

func newNoopAuthorizer(t *tokenChecker) *NoopAuthorizer {
	return &NoopAuthorizer{t: t}
}

func (n *NoopAuthorizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeSessionResponse(w, n.t)
}

func (*NoopAuthorizer) TypeName() string { return "none" }

type BasicAuthorizer struct {
	username string
	password string
	realm    string
	t        *tokenChecker
}

type basicAuthorizerConfig struct {
	Username string
	Password string
	Realm    string
}

func newBasicAuthorizer(t *tokenChecker, cfg basicAuthorizerConfig) *BasicAuthorizer {
	return &BasicAuthorizer{
		username: cfg.Username,
		password: cfg.Password,
		realm:    cfg.Realm,
		t:        t,
	}
}

func (b *BasicAuthorizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, isBasicAuth := r.BasicAuth()
	if !isBasicAuth {
		w.Header().Add("WWW-Authenticate", b.realm)
		writeResponse(w, http.StatusUnauthorized, APIError{Error: "No valid basic auth provided"})
		return
	}
	if username != b.username || password != b.password {
		writeResponse(w, http.StatusForbidden, APIError{Error: "Incorrect username or password"})
		return
	}
	writeSessionResponse(w, b.t)
}

func (*BasicAuthorizer) TypeName() string { return "basic" }

type tokenChecker struct {
	sessions   map[string]time.Time
	logger     *log.Logger
	sessionLen time.Duration
	lock       sync.Mutex
}

type tokenCheckerConfig struct {
	Logger          *log.Logger
	SessionDuration time.Duration
}

func newTokenChecker(cfg tokenCheckerConfig) *tokenChecker {
	return &tokenChecker{
		sessions:   make(map[string]time.Time),
		logger:     cfg.Logger,
		sessionLen: cfg.SessionDuration,
	}
}

func (t *tokenChecker) newSession() (string, error) {
	randomValue := make([]byte, 16)
	_, err := rand.Read(randomValue)
	if err != nil {
		t.logger.Error("Could not generate random bits for auth session token: %s", err)
		return "", err
	}

	sessionToken := base64.RawStdEncoding.EncodeToString(randomValue)
	expiryTime := time.Now().Add(t.sessionLen)
	t.lock.Lock()
	t.sessions[sessionToken] = expiryTime
	t.lock.Unlock()
	return sessionToken, nil
}

func (t *tokenChecker) validate(token string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	requestTime := time.Now()
	expireTime, found := t.sessions[token]
	if !found {
		return false
	}

	if expireTime.Before(requestTime) {
		delete(t.sessions, token)
		return false
	}

	t.sessions[token] = requestTime.Add(t.sessionLen)
	return true
}

func (t *tokenChecker) wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sessionToken := r.Header.Get("Signalfire-Session")
			if !t.validate(sessionToken) {
				writeResponse(w, http.StatusUnauthorized, APIError{
					Error: "Auth token invalid",
				})
				return
			}

			h.ServeHTTP(w, r)
		},
	)
}
