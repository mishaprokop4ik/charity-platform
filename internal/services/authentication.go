package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"math/rand"
	"time"
)

type Authentication struct {
	repo       *repository.Repository
	authConfig *config.AuthenticationConfig
}

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

func NewAuthentication(repo *repository.Repository, authConfig *config.AuthenticationConfig) *Authentication {
	return &Authentication{repo: repo, authConfig: authConfig}
}

func (a *Authentication) SignUp(ctx context.Context, user models.User) (uint, error) {
	user.Password = GeneratePasswordHash(user.Password, a.authConfig.Salt)
	return a.repo.User.CreateUser(ctx, user)
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

type TokenClaims struct {
	jwt.StandardClaims
	ID uint `json:"id"`
}
