package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type AuthenticationConfig struct {
	Salt                 string `yaml:"salt"`
	SigningKey           string `yaml:"signingKey"`
	TokenExpirationHours uint   `yaml:"tokenExpirationHours"`
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
