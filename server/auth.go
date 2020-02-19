package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/starkandwayne/signalfire/config"
	"github.com/starkandwayne/signalfire/log"
)

const (
	AuthCookieName = "signalfire-session"
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
	case "userpass":
		auth = newUserpassAuthorizer(t, userpassAuthorizerConfig{
			Username: conf.Username,
			Password: conf.Password,
		})
	default:
		err = fmt.Errorf("Unknown auth type: %s", conf.Type)
	}

	return
}

func writeSessionResponse(w http.ResponseWriter, t *tokenChecker) {
	sessionToken, expiry, err := t.newSession()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, APIError{Error: err.Error()})
	}
	http.SetCookie(w, &http.Cookie{
		Name:    AuthCookieName,
		Value:   sessionToken,
		Path:    "/",
		Expires: expiry,
	})
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

type UserpassAuthorizer struct {
	username string
	password string
	t        *tokenChecker
}

type userpassAuthorizerConfig struct {
	Username string
	Password string
}

func newUserpassAuthorizer(t *tokenChecker, cfg userpassAuthorizerConfig) *UserpassAuthorizer {
	return &UserpassAuthorizer{
		username: cfg.Username,
		password: cfg.Password,
		t:        t,
	}
}

type UserpassAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b *UserpassAuthorizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestParameters := UserpassAuthRequest{}
	jsonDec := json.NewDecoder(r.Body)
	err := jsonDec.Decode(&requestParameters)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, APIError{Error: "JSON request body could not be parsed"})
		return
	}

	if requestParameters.Username != b.username || requestParameters.Password != b.password {
		writeResponse(w, http.StatusForbidden, APIError{Error: "Incorrect username or password"})
		return
	}

	writeSessionResponse(w, b.t)
}

func (*UserpassAuthorizer) TypeName() string { return "userpass" }

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

func (t *tokenChecker) newSession() (string, time.Time, error) {
	randomValue := make([]byte, 16)
	_, err := rand.Read(randomValue)
	if err != nil {
		t.logger.Error("Could not generate random bits for auth session token: %s", err)
		return "", time.Time{}, err
	}

	sessionToken := base64.RawStdEncoding.EncodeToString(randomValue)
	expiryTime := time.Now().Add(t.sessionLen)
	t.lock.Lock()
	t.sessions[sessionToken] = expiryTime
	t.lock.Unlock()
	return sessionToken, expiryTime, nil
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
			if sessionToken == "" {
				cookie, err := r.Cookie(AuthCookieName)
				if err == nil {
					sessionToken = cookie.Value
				}
			}

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
