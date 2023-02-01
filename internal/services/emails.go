package service

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Sender struct {
	email, password string
	SMTPEndpoint    string
}

func (s *Sender) SendEmail(receive, body, emailType string) error {
	mimeType := ""
	switch emailType {
	case "html":
		mimeType = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	default:
		return fmt.Errorf("incorrect email type")
	}
	subject := fmt.Sprintf("Subject: Confirming email %s for Kurajj charity platform!\n", receive)
	msg := []byte(subject + mimeType + body)
	host, err := s.getSMPTHost()

	if err != nil {
		return err
	}
	err = smtp.SendMail(s.SMTPEndpoint,
		smtp.PlainAuth("", s.email, s.password, host),
		s.email, []string{receive}, msg)

	return err
}

func (s *Sender) getSMPTHost() (string, error) {
	smptEndpointData := strings.Split(s.SMTPEndpoint, ":")
	if len(smptEndpointData) != 2 {
		return "", fmt.Errorf("format of smtp endpoint is incorrect: wanted: `host:port`, got: %s", s.SMTPEndpoint)
	}

	return smptEndpointData[0], nil
}
