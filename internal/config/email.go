package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Email struct {
	Email        string `yaml:"email"`
	Password     string `yaml:"password"`
	SMPTEndpoint string `yaml:"smptEndpoint"`
}

func NewEmailConfigFromFile(filename string) (Email, error) {
	configData, err := os.ReadFile(filename)
	if err != nil {
		return Email{}, err
	}
	c := Email{}
	if err := yaml.Unmarshal(configData, &c); err != nil {
		return Email{}, err
	}

	return c, nil
}
