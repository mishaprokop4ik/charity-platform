package configs

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type AuthenticationConfig struct {
	Salt            string        `yaml:"salt"`
	SigningKey      string        `yaml:"signingKey"`
	AccessTokenTTL  time.Duration `yaml:"accessTokenTTL"`
	RefreshTokenTTL time.Duration `yaml:"refreshTokenTTL"`
	Key             string        `yaml:"key"`
}

func NewAuthenticationConfigFromFile(filename string) (AuthenticationConfig, error) {
	configData, err := os.ReadFile(filename)
	if err != nil {
		return AuthenticationConfig{}, err
	}
	c := AuthenticationConfig{}
	if err := yaml.Unmarshal(configData, &c); err != nil {
		return AuthenticationConfig{}, err
	}

	return c, nil
}
