package token

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type UserClaims struct {
	Username  string    `json:"username"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewUserClaims(username string, nickname string, duration time.Duration) (*UserClaims, error) {
	claims := &UserClaims{
		Username:  username,
		ExpiredAt: time.Now().Add(duration),
	}
	return claims, nil
}

func (claim *UserClaims) Valid() error {
	if time.Now().After(claim.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
