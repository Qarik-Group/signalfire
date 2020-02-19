package config

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

type Config struct {
	Targets []BOSH `yaml:"targets"`
	Server  Server `yaml:"server"`
	Log     Log    `yaml:"log"`
}
type BOSH struct {
	URL                string `yaml:"url"`
	CACert             string `yaml:"ca_cert"`
	PollInterval       uint   `yaml:"poll_interval"` //in seconds
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	Auth               struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"auth"`
}

type Server struct {
	TLS struct {
		Certificate string `yaml:"certificate"`
		PrivateKey  string `yaml:"private_key"`
	} `yaml:"tls"`
	Auth Auth   `yaml:"auth"`
	Port uint16 `yaml:"port"`
}

type Auth struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Log struct {
	Level string `yaml:"level"`
}

var DefaultConfig = Config{
	Server: Server{
		Port: 11001,
		Auth: Auth{
			Type:     "userpass",
			Username: "admin",
			Password: "password",
		},
	},
	Log: Log{Level: "info"},
}

func Parse(r io.Reader) (*Config, error) {
	d := yaml.NewDecoder(r)
	ret := DefaultConfig
	err := d.Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("Error decoding config yaml: %s", err)
	}
	for _, t := range ret.Targets {
		if t.PollInterval == 0 {
			t.PollInterval = 30
		}
	}
	ret.Log.Level = strings.ToLower(ret.Log.Level)
	return &ret, nil
}
