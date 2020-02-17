package bosh

import (
	"encoding/base64"

	"github.com/starkandwayne/signalfire/log"
)

func (b *Client) authHeader() string {
	var ret string
	b.authLock.RLock()
	if b.auth != nil {
		ret = b.auth.Header()
	}
	b.authLock.RUnlock()
	return ret
}

func (b *Client) login() error {
	b.authLock.Lock()
	err := b.auth.Login()
	b.authLock.Unlock()
	return err
}

type boshAuthorizer interface {
	Login() error
	Header() string
}

type basicAuth struct {
	Username string
	Password string
}

func (b *basicAuth) Login() error { return nil }

func (b *basicAuth) Header() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(b.Username+":"+b.Password))
}

type uaaAuth struct {
	Client      *UAA
	Logger      *log.Logger
	accessToken string
	Username    string
	Password    string
}

func (u *uaaAuth) Login() error {
	resp, err := u.Client.ClientCredentials(u.Username, u.Password)
	if err != nil {
		return err
	}

	u.accessToken = resp.AccessToken
	return nil
}

func (u *uaaAuth) Header() string {
	return "Bearer " + u.accessToken
}
