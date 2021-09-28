package token

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenMaker struct {
	secretKey string
}

func NewTokenMaker(secretKey string) (*TokenMaker, error) {
	return &TokenMaker{secretKey}, nil
}

func (maker *TokenMaker) CreateToken(username string, nickname string, duration time.Duration) (string, error) {
	payload, err := NewUserClaims(username, nickname, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *TokenMaker) VerifyToken(token string) (*UserClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &UserClaims{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := jwtToken.Claims.(*UserClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
