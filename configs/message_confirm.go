package configs

import (
	"gopkg.in/yaml.v3"
	"os"
)

type MessageConfirm struct {
	Account     string `json:"account,omitempty" yaml:"account"`
	Password    string `json:"password,omitempty" yaml:"password"`
	PhoneNumber string `json:"phoneNumber,omitempty" yaml:"phoneNumber"`
}

func NewMessageConfirm(filename string) (MessageConfirm, error) {
	configData, err := os.ReadFile(filename)
	if err != nil {
		return MessageConfirm{}, err
	}
	c := MessageConfirm{}
	if err := yaml.Unmarshal(configData, &c); err != nil {
		return MessageConfirm{}, err
	}

	return c, nil
}
