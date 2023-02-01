package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	zlog "Kurajj/pkg/logger"
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"html/template"
	"math/rand"
	"time"
)

type Authenticator interface {
	SignUp(ctx context.Context, user models.User) (uint, error)
	SignIn(ctx context.Context, user models.User) (models.SignedInUser, error)
	ParseToken(accessToken string) (uint, error)
	ConfirmEmail(ctx context.Context, email string) error
}

type Authentication struct {
	repo        *repository.Repository
	authConfig  *config.AuthenticationConfig
	emailSender Sender
}

const confirmEmail = "Confirm Your Email Address"

func GenerateRandomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	length := 10 + rand.Intn(5)
	s := make([]byte, length)
	for i := range s {
		s[i] = charset[rand.Intn(len(charset))]
	}
	return string(s)
}

func NewAuthentication(repo *repository.Repository, authConfig *config.AuthenticationConfig, emailConfig *config.Email) *Authentication {
	return &Authentication{repo: repo, authConfig: authConfig, emailSender: Sender{
		email:        emailConfig.Email,
		password:     emailConfig.Password,
		SMTPEndpoint: emailConfig.SMPTEndpoint,
	}}
}

type EmailCheck struct {
	Title         string
	Email         string
	ServerConfirm string
}

func (a *Authentication) SignUp(ctx context.Context, user models.User) (uint, error) {
	user.Password = GeneratePasswordHash(user.Password, a.authConfig.Salt)
	id, err := a.repo.User.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	confirmEmailBody := bytes.Buffer{}

	confirmEmailValues := EmailCheck{
		Title:         confirmEmail,
		Email:         user.Email,
		ServerConfirm: "http://localhost:8080/auth/confirm",
	}

	confirmEmailTmpl, err := template.New("confirm_email.tmpl").ParseFiles("internal/templates/confirm_email.tmpl")
	if err != nil {
		zlog.Log.Error(err, "could not parse confirm_email.tmpl")
		return 0, err
	}

	err = confirmEmailTmpl.Execute(&confirmEmailBody, confirmEmailValues)
	if err != nil {
		zlog.Log.Error(err, "could not create confirm email body")
		return 0, err
	}

	err = a.emailSender.SendEmail(user.Email, confirmEmailBody.String(), "html")
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GeneratePasswordHash(password, salt string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func (a *Authentication) generateToken(ctx context.Context, email, password string) (string, error) {
	userID, err := a.repo.User.GetUserAuthentication(ctx, email, GeneratePasswordHash(password, a.authConfig.Salt))
	if err != nil {
		return "", err
	}
	expirationAfterHours := time.Duration(a.authConfig.TokenExpirationHours) * time.Hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationAfterHours).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
	})
	resp, err := token.SignedString([]byte(a.authConfig.SigningKey))
	return resp, err
}

func (a *Authentication) ParseToken(accessToken string) (uint, error) {
	token, err := jwt.ParseWithClaims(accessToken, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}

		return []byte(a.authConfig.SigningKey), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type TokenClaims")
	}

	return claims.ID, nil
}

func (a *Authentication) SignIn(ctx context.Context, user models.User) (models.SignedInUser, error) {
	token, err := a.generateToken(ctx, user.Email, user.Password)
	if err != nil {
		return models.SignedInUser{}, err
	}
	user.Password = GeneratePasswordHash(user.Password, a.authConfig.Salt)
	userInformation, err := a.repo.User.GetEntity(ctx, user.Email, user.Password, user.IsAdmin, false)
	if err != nil {
		return models.SignedInUser{}, err
	}
	return userInformation.GetUserFullResponse(token), nil
}

func (a *Authentication) ConfirmEmail(ctx context.Context, email string) error {
	userValues := models.User{
		Email:       email,
		IsActivated: true,
	}.GetValuesToUpdate()

	return a.repo.User.UpdateUserByEmail(ctx, email, userValues)
}

type TokenClaims struct {
	jwt.StandardClaims
	ID uint `json:"id"`
}
