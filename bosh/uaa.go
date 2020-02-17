package bosh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/starkandwayne/signalfire/log"
)

type UAA struct {
	URL    string
	Client *http.Client
	Logger *log.Logger
}

func (c *UAA) do(values url.Values) (*UAAAuthResponse, error) {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/oauth/token", c.URL),
		strings.NewReader(values.Encode()),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	c.Logger.Debug(string(reqDump))
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	c.Logger.Debug(string(respDump))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not authenticate: Status %d", resp.StatusCode)
	}

	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	r := response{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&r)
	if err != nil {
		return nil, err
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &UAAAuthResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TTL:          time.Second * time.Duration(r.ExpiresIn),
	}, nil
}

type UAAAuthResponse struct {
	AccessToken  string
	RefreshToken string
	TTL          time.Duration
}

func (c *UAA) ClientCredentials(
	clientID,
	clientSecret string) (*UAAAuthResponse, error) {

	return c.do(url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{clientID},
		"client_secret": []string{clientSecret},
	})
}
