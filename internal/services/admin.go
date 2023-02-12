package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	zlog "Kurajj/pkg/logger"
	"bytes"
	"context"
	"html/template"
)

func NewAdmin(repo *repository.Repository, authConfig *config.AuthenticationConfig, emailConfig *config.Email) *Admin {
	return &Admin{repo: repo, authConfig: authConfig, emailSender: Sender{
		email:        emailConfig.Email,
		password:     emailConfig.Password,
		SMTPEndpoint: emailConfig.SMPTEndpoint,
	}}
}

type Admin struct {
	repo        *repository.Repository
	authConfig  *config.AuthenticationConfig
	emailSender Sender
}

type OneTimePassword struct {
	Password string
}

func (a *Admin) CreateAdmin(ctx context.Context, admin models.User) (uint, error) {
	password := GenerateRandomPassword()
	admin.Password = GeneratePasswordHash(password, a.authConfig.Salt)
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
	return a.repo.Admin.CreateAdmin(ctx, admin)
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

type AdminCRUDer interface {
	CreateAdmin(ctx context.Context, admin models.User) (uint, error)
	GetAdminByID(ctx context.Context, id uint) (models.User, error)
	UpdateAdmin(ctx context.Context, admin models.User) error
	DeleteAdmin(ctx context.Context, id uint) error
	GetAllAdmins(ctx context.Context) ([]models.User, error)
}

const oneTimePasswordEmail = "One time password"
