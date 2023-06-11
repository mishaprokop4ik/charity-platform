package service

import (
	"Kurajj/configs"
	"Kurajj/internal/models"
	"Kurajj/pkg/encrypt"
	"Kurajj/pkg/hash"
	zlog "Kurajj/pkg/logger"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
)

func NewAdmin(repo Repositorier, authConfig *configs.AuthenticationConfig, emailConfig *configs.Email) *Admin {
	return &Admin{repo: repo, authConfig: authConfig, emailSender: Sender{
		email:        emailConfig.Email,
		password:     emailConfig.Password,
		SMTPEndpoint: emailConfig.SMPTEndpoint,
	}}
}

type Admin struct {
	repo        Repositorier
	authConfig  *configs.AuthenticationConfig
	emailSender Sender
}

type OneTimePassword struct {
	Password string
}

func (a *Admin) CreateAdmin(ctx context.Context, admin models.User) (uint, error) {
	password := GenerateRandomPassword()
	admin.Password = hash.GenerateHash(password, a.authConfig.Salt)
	admin.SearchIndex = hash.GenerateHash(admin.Email, a.authConfig.Salt)
	oneTimePasswordBody := bytes.Buffer{}

	oneTimePasswordValues := OneTimePassword{
		Password: password,
	}

	oneTimePasswordTmpl, err := template.New("one_time_password_email.tmpl").ParseFiles("internal/templates/one_time_password_email.tmpl")
	if err != nil {
		zlog.Log.Error(err, "could not parse one_time_password_email.tmpl")
		return 0, err
	}

	err = oneTimePasswordTmpl.Execute(&oneTimePasswordBody, oneTimePasswordValues)
	if err != nil {
		zlog.Log.Error(err, "could not create confirm email body")
		return 0, err
	}

	err = a.emailSender.SendEmail(admin.Email, oneTimePasswordBody.String(), "html")
	if err != nil {
		return 0, err
	}
	if err := a.encryptUserPersonalData(&admin); err != nil {
		zlog.Log.Error(err, "admin cannot be created, because encryption failed")
		return 0, fmt.Errorf("cannot encrypt user sensetive fields, %v", err)
	}
	return a.repo.CreateAdmin(ctx, admin)
}

func (a *Admin) encryptUserPersonalData(admin *models.User) error {
	signingKey := a.authConfig.Key
	encryptedEmail, err := encrypt.Encrypt(admin.Email, signingKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt email: %v", err)
	}
	admin.Email = encryptedEmail
	if phone := strings.TrimSpace(admin.Telephone); phone != "" {
		encryptedTelephone, err := encrypt.Encrypt(phone, signingKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt email: %v", err)
		}
		admin.Telephone = encryptedTelephone
	}

	if telegramUsername := strings.TrimSpace(admin.TelegramUsername); telegramUsername != "" {
		encryptedTelegramUsername, err := encrypt.Encrypt(telegramUsername, signingKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt email: %v", err)
		}
		admin.TelegramUsername = encryptedTelegramUsername
	}

	return nil
}

func (a *Admin) GetAdminByID(ctx context.Context, id uint) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Admin) UpdateAdmin(ctx context.Context, admin models.User) error {
	//TODO implement me
	panic("implement me")
}

func (a *Admin) DeleteAdmin(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}

func (a *Admin) GetAllAdmins(ctx context.Context) ([]models.User, error) {
	//TODO implement me
	panic("implement me")
}
