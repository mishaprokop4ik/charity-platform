package service

import (
	"Kurajj/configs"
	"Kurajj/internal/models"
	"Kurajj/pkg/encrypt"
	"Kurajj/pkg/hash"
	zlog "Kurajj/pkg/logger"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/twilio/twilio-go"
	twilloAPI "github.com/twilio/twilio-go/rest/api/v2010"
	"gorm.io/gorm"
	"html/template"
	"math/rand"
	"strings"
	"time"
)

type Authenticator interface {
	GetUserShortInfo(ctx context.Context, id uint) (models.UserShortInfo, error)
	SignUp(ctx context.Context, user models.User) (uint, error)
	SignIn(ctx context.Context, user models.User) (models.SignedInUser, error)
	GetUserByRefreshToken(ctx context.Context, token string) (models.SignedInUser, error)
	ParseToken(accessToken string) (uint, error)
	NewRefreshToken() (string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (models.Tokens, error)
	ConfirmEmail(ctx context.Context, email string) error
	UpdateEntity(ctx context.Context, entity models.UserUpdate) error
	SendMessage(ctx context.Context, message models.ConfirmMessage) error
	ConfirmUserByPhoneCode(ctx context.Context, confirm models.UserConfirm) error
}

func NewAuthentication(repo Repositorier,
	authConfig *configs.AuthenticationConfig,
	emailConfig *configs.Email,
	messageConfig *configs.MessageConfirm) *Authentication {
	return &Authentication{repo: repo, authConfig: authConfig, emailSender: Sender{
		email:        emailConfig.Email,
		password:     emailConfig.Password,
		SMTPEndpoint: emailConfig.SMPTEndpoint,
	}, VerificationPhoneNumber: messageConfig.PhoneNumber, MessageAuthToken: messageConfig.Password, AccountSid: messageConfig.Account}
}

type Authentication struct {
	repo                         Repositorier
	authConfig                   *configs.AuthenticationConfig
	emailSender                  Sender
	VerificationPhoneNumber      string
	MessageAuthToken, AccountSid string
}

func (a *Authentication) ConfirmUserByPhoneCode(ctx context.Context, confirm models.UserConfirm) error {
	user, err := a.repo.GetUserInfo(ctx, uint(confirm.UserID))
	if err != nil {
		return err
	}

	if len(confirm.ConfirmCode) != len(user.ConfirmCode) {
		return fmt.Errorf("incorrect code size")
	}

	for i := range user.ConfirmCode {
		if user.ConfirmCode[i] != confirm.ConfirmCode[i] {
			return fmt.Errorf("incorrect code")
		}
	}

	user.IsActivated = true
	err = a.repo.UpdateUser(ctx, models.UserUpdate{
		Model: gorm.Model{
			ID: user.ID,
		},
		IsActivated: &user.IsActivated,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Authentication) SendMessage(ctx context.Context, message models.ConfirmMessage) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   a.AccountSid,
		Password:   a.MessageAuthToken,
		AccountSid: a.AccountSid,
	})

	params := &twilloAPI.CreateMessageParams{}
	params.SetBody(message.Text)
	params.SetFrom(a.VerificationPhoneNumber)
	params.SetTo(message.To)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		zlog.Log.Error(err, "could not send message")
		return err
	}
	if resp.Status != nil {
		zlog.Log.Info(*resp.Status, "phone number", message.To)
	}

	return nil
}

func (a *Authentication) GetUserShortInfo(ctx context.Context, id uint) (models.UserShortInfo, error) {
	fullUser, err := a.repo.GetUserInfo(ctx, id)
	if err != nil {
		return models.UserShortInfo{}, err
	}

	user := models.UserShortInfo{
		ID:              fullUser.ID,
		Username:        fullUser.FullName,
		ProfileImageURL: fullUser.AvatarImagePath,
	}

	return user, nil
}

func (a *Authentication) GetUserByRefreshToken(ctx context.Context, token string) (models.SignedInUser, error) {
	user, err := a.repo.GetByRefreshToken(ctx, token)
	if err != nil {
		return models.SignedInUser{}, err
	}

	return user.GetUserFullResponse(models.Tokens{
		Access:  "",
		Refresh: user.RefreshToken,
	}), nil
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

type EmailCheck struct {
	Title         string
	Email         string
	ServerConfirm string
}

func (a *Authentication) SignUp(ctx context.Context, user models.User) (uint, error) {
	isEmailTaken, err := a.repo.IsEmailTaken(ctx, user.Email)
	if err != nil {
		return 0, err
	}
	if isEmailTaken {
		return 0, fmt.Errorf("email %s is taken", user.Email)
	}
	user.Password = hash.GenerateHash(user.Password, a.authConfig.Salt)
	user.SearchIndex = hash.GenerateHash(user.Email, a.authConfig.Salt)
	code := make([]int64, 6)
	for i := range code {
		code[i] = int64(rand.Intn(9))
	}

	user.ConfirmCode = code

	emailPage, err := a.generateEmail(user.Email)
	if err != nil {
		return 0, err
	}

	err = a.emailSender.SendEmail(user.Email, emailPage.String(), "html")
	if err != nil {
		return 0, err
	}

	err = a.SendMessage(ctx, models.ConfirmMessage{
		Text: fmt.Sprintf("Hi %s, please confirm your account by entering next Code: %v", user.FullName, code),
		To:   user.Telephone,
	})

	if err := a.encryptUserPersonalData(&user); err != nil {
		zlog.Log.Error(err, "user cannot be created, because encryption failed")
		return 0, fmt.Errorf("cannot encrypt user sensetive fields, %v", err)
	}

	id, err := a.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a *Authentication) encryptUserPersonalData(user *models.User) error {
	signingKey := a.authConfig.Key
	encryptedEmail, err := encrypt.Encrypt(user.Email, signingKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt email: %v", err)
	}
	user.Email = encryptedEmail
	if phone := strings.TrimSpace(user.Telephone); phone != "" {
		encryptedTelephone, err := encrypt.Encrypt(phone, signingKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt email: %v", err)
		}
		user.Telephone = encryptedTelephone
	}

	if telegramUsername := strings.TrimSpace(user.TelegramUsername); telegramUsername != "" {
		encryptedTelegramUsername, err := encrypt.Encrypt(telegramUsername, signingKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt email: %v", err)
		}
		user.TelegramUsername = encryptedTelegramUsername
	}

	return nil
}

func (a *Authentication) generateEmail(email string) (fmt.Stringer, error) {
	confirmEmailBody := bytes.Buffer{}

	confirmEmailValues := EmailCheck{
		Title:         confirmEmail,
		Email:         email,
		ServerConfirm: "http://localhost:8080/auth/confirm",
	}

	confirmEmailTmpl, err := template.New("confirm_email.tmpl").ParseFiles("internal/templates/confirm_email.tmpl")
	if err != nil {
		zlog.Log.Error(err, "could not parse confirm_email.tmpl")
		return nil, err
	}

	err = confirmEmailTmpl.Execute(&confirmEmailBody, confirmEmailValues)
	if err != nil {
		zlog.Log.Error(err, "could not create confirm email body")
		return nil, err
	}

	return &confirmEmailBody, nil
}

func (a *Authentication) generateAccessToken(_ context.Context, userID uint, isAdmin bool) (string, error) {
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
	user.Password = hash.GenerateHash(user.Password, a.authConfig.Salt)
	user.SearchIndex = hash.GenerateHash(user.Email, a.authConfig.Salt)

	userInformation, err := a.repo.GetEntity(ctx, user.SearchIndex, user.Password, user.IsAdmin, false)
	if err != nil {
		return models.SignedInUser{}, err
	}
	tokens, err := a.createSession(ctx, userInformation.ID, userInformation.IsAdmin)
	if err != nil {
		return models.SignedInUser{}, err
	}
	fmt.Println(userInformation.Email, "there1")
	if err := a.decryptUserPersonalData(&userInformation); err != nil {
		return models.SignedInUser{}, err
	}
	return userInformation.GetUserFullResponse(tokens), nil
}

func (a *Authentication) decryptUserPersonalData(user *models.User) error {
	signingKey := a.authConfig.Key
	fmt.Println(user.Email, "there2")
	decryptedEmail, err := encrypt.Decrypt(user.Email, signingKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt email: %v", err)
	}
	user.Email = decryptedEmail
	if phone := strings.TrimSpace(user.Telephone); phone != "" {
		decryptedTelephone, err := encrypt.Decrypt(phone, signingKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt email: %v", err)
		}
		user.Telephone = decryptedTelephone
	}

	if telegramUsername := strings.TrimSpace(user.TelegramUsername); telegramUsername != "" {
		decryptedTelegramUsername, err := encrypt.Decrypt(telegramUsername, signingKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt email: %v", err)
		}
		user.TelegramUsername = decryptedTelegramUsername
	}

	return nil
}

func (a *Authentication) ConfirmEmail(ctx context.Context, email string) error {
	userValues := models.User{
		Email:       email,
		IsActivated: true,
	}.GetValuesToUpdate()

	return a.repo.UpdateUserByEmail(ctx, email, userValues)
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
	member, err := a.repo.GetByRefreshToken(ctx, refreshToken)
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

	err = a.repo.SetSession(ctx, userID, session)

	return res, err
}

func (a *Authentication) UpdateEntity(ctx context.Context, entity models.UserUpdate) error {
	if entity.Email != nil {
		emailPage, err := a.generateEmail(*entity.Email)
		if err != nil {
			return err
		}
		err = a.emailSender.SendEmail(*entity.Email, emailPage.String(), "html")
		if err != nil {
			return err
		}
	}
	return a.repo.UpdateUser(ctx, entity)
}

type TokenClaims struct {
	jwt.StandardClaims
	ID      uint `json:"id"`
	IsAdmin bool `json:"isAdmin"`
}
