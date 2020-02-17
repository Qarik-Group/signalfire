package bosh

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/starkandwayne/signalfire/config"
	"github.com/starkandwayne/signalfire/log"
)

type Client struct {
	client   *http.Client
	logger   *log.Logger
	url      string
	auth     boshAuthorizer
	name     string
	uuid     string
	username string
	password string
	authLock sync.RWMutex
}

func NewClient(config config.BOSH, logger *log.Logger) (*Client, error) {
	logger.Debug("Initializing BOSH client")
	certPool, err := certPoolFrom(config.CACert)
	if err != nil {
		return nil, fmt.Errorf("Cannot initialize cert pool: %s", err)
	}

	u, err := canonizeURL(config.URL)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            certPool,
					InsecureSkipVerify: config.InsecureSkipVerify,
				},
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
			},
		},
		username: config.Auth.ClientID,
		password: config.Auth.ClientSecret,
		logger:   logger,
		url:      u,
	}, nil
}

func canonizeURL(uStr string) (string, error) {
	var schemeRegex = regexp.MustCompile("^(http|https)://")
	if !schemeRegex.MatchString(uStr) {
		uStr = "https://" + uStr
	}

	u, err := url.Parse(uStr)
	if err != nil {
		return "", fmt.Errorf("Error parsing BOSH URL: %s", err)
	}

	if u.Port() == "" {
		u.Host = u.Host + ":25555"
	}

	return strings.TrimSuffix(u.String(), "/"), nil
}

func certPoolFrom(cert string) (*x509.CertPool, error) {
	if cert == "" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("Error fetching system cert pool: %s", err)
		}
		return pool, nil
	}

	p, _ := pem.Decode([]byte(cert))
	if p == nil {
		return nil, fmt.Errorf("Could not parse certificate input as PEM")
	}

	c, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Could not parse certificate ASN.1 data: %s", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(c)
	return pool, nil
}

func (b *Client) Connect() error {
	info, err := b.info()
	if err != nil {
		return err
	}

	b.name = info.Name
	b.uuid = info.UUID

	switch info.Auth.Type {
	case "basic":
		b.auth = &basicAuth{
			Username: b.username,
			Password: b.password}
	case "uaa":
		b.auth = &uaaAuth{
			Client: &UAA{
				URL:    info.Auth.Options.URL,
				Client: b.client,
				Logger: b.logger,
			},
			Logger:   b.logger,
			Username: b.username,
			Password: b.password,
		}
	}

	err = b.login()
	if err != nil {
		return fmt.Errorf("Error when logging in for the first time: %s", err)
	}

	go func() {
		for range time.Tick(30 * time.Second) {
			b.logger.Debug("Triggering BOSH authentication")
			err := b.login()
			if err != nil {
				b.logger.Error("Error when logging in: %s", err)
			}
		}
	}()

	return nil
}

func (b *Client) do(req *http.Request, output interface{}) error {
	dump, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		b.logger.Debug("%s", string(dump))
	}

	req.Header.Add("Authorization", b.authHeader())

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	dump, err = httputil.DumpResponse(resp, true)
	if err == nil {
		b.logger.Debug("%s", string(dump))
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf(resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if output != nil {
		err := json.Unmarshal(bodyBytes, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Client) path(path string) string {
	return b.url + "/" + strings.TrimPrefix(path, "/")
}

type infoOut struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Auth struct {
		Type    string `json:"type"`
		Options struct {
			URL string `json:"url"`
		} `json:"options"`
	} `json:"user_authentication"`
}

func (b *Client) info() (*infoOut, error) {
	req, err := http.NewRequest("GET", b.path("/info"), nil)
	if err != nil {
		return nil, err
	}

	info := infoOut{}
	err = b.do(req, &info)
	if err != nil {
		return nil, fmt.Errorf("Error getting info: %s", err)
	}

	return &info, nil
}

type Deployment struct {
	Name     string    `json:"name"`
	Releases []Release `json:"releases"`
}

type Release struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (b *Client) Deployments() ([]Deployment, error) {
	req, err := http.NewRequest("GET", b.path("/deployments"), nil)
	if err != nil {
		return nil, err
	}

	ret := []Deployment{}
	err = b.do(req, &ret)
	if err != nil {
		return nil, fmt.Errorf("Error getting deployments: %s", err)
	}

	return ret, nil
}

func (b *Client) Name() string { return b.name }
func (b *Client) UUID() string { return b.uuid }
