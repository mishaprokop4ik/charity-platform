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
	GetUserShortInfo(ctx context.Context, id uint) (models.UserComment, error)
	SignUp(ctx context.Context, user models.User) (uint, error)
	SignIn(ctx context.Context, user models.User) (models.SignedInUser, error)
	ParseToken(accessToken string) (uint, error)
	NewRefreshToken() (string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (models.Tokens, error)
	ConfirmEmail(ctx context.Context, email string) error
}

type Authentication struct {
	repo        *repository.Repository
	authConfig  *config.AuthenticationConfig
	emailSender Sender
}

func (a *Authentication) GetUserShortInfo(ctx context.Context, id uint) (models.UserComment, error) {
	fullUser, err := a.repo.User.GetUserInfo(ctx, id)
	if err != nil {
		return models.UserComment{}, err
	}

	user := models.UserComment{
		AuthorID:        fullUser.ID,
		Username:        fullUser.FullName,
		ProfileImageURL: fullUser.AvatarImagePath,
	}

	return user, nil
}

const confirmEmail = "Confirm Your Email Address"

func GenerateRandomPassword() string {
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
	//isEmailTaken, err := a.repo.User.IsEmailTaken(ctx, user.Email)
	//if err != nil {
	//	return 0, err
	//}
	//if isEmailTaken {
	//	return 0, fmt.Errorf("email %s is taken", user.Email)
	//}
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

func (a *Authentication) generateAccessToken(ctx context.Context, userID uint, isAdmin bool) (string, error) {
	expirationAfterHours := a.authConfig.AccessTokenTTL
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationAfterHours).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
		isAdmin,
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
		return 0, errors.New("token claims are incorrect")
	}

	return claims.ID, nil
}

func (a *Authentication) SignIn(ctx context.Context, user models.User) (models.SignedInUser, error) {
	user.Password = GeneratePasswordHash(user.Password, a.authConfig.Salt)
	userInformation, err := a.repo.User.GetEntity(ctx, user.Email, user.Password, user.IsAdmin, false)
	if err != nil {
		return models.SignedInUser{}, err
	}
	tokens, err := a.createSession(ctx, userInformation.ID, userInformation.IsAdmin)
	if err != nil {
		return models.SignedInUser{}, err
	}
	return userInformation.GetUserFullResponse(tokens), nil
}

func (a *Authentication) ConfirmEmail(ctx context.Context, email string) error {
	userValues := models.User{
		Email:       email,
		IsActivated: true,
	}.GetValuesToUpdate()

	return a.repo.User.UpdateUserByEmail(ctx, email, userValues)
}

func (a *Authentication) NewRefreshToken() (string, error) {
	token := make([]byte, 32)

	sourceSeeder := rand.NewSource(time.Now().Unix())
	randomGenerator := rand.New(sourceSeeder)

	_, err := randomGenerator.Read(token)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", token), nil
}

func (a *Authentication) RefreshTokens(ctx context.Context, refreshToken string) (models.Tokens, error) {
	member, err := a.repo.User.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return models.Tokens{}, err
	}
	zlog.Log.Info("got member", "id", member.ID)
	return a.createSession(ctx, member.ID, member.IsAdmin)
}

func (a *Authentication) createSession(ctx context.Context, userID uint, isAdmin bool) (models.Tokens, error) {
	var (
		res models.Tokens
		err error
	)

	res.Access, err = a.generateAccessToken(ctx, userID, isAdmin)
	if err != nil {
		return res, err
	}

	res.Refresh, err = a.NewRefreshToken()
	if err != nil {
		return res, err
	}

	session := models.MemberSession{
		RefreshToken: res.Refresh,
		ExpiresAt:    time.Now().Add(a.authConfig.RefreshTokenTTL),
	}

	err = a.repo.User.SetSession(ctx, userID, session)

	return res, err
}

type TokenClaims struct {
	jwt.StandardClaims
	ID      uint `json:"id"`
	IsAdmin bool `json:"isAdmin"`
}
