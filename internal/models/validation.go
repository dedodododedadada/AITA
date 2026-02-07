package models

import (
	"aita/internal/pkg/utils"
	"strings"
	"unicode/utf8"
)

func IsValidSignUpReq(req *SignupRequest) error {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return ErrRequiredFieldMissing
	}
	if len(req.Username) < 4 || len(req.Username) > 50 {
		return ErrInvalidUsernameFormat
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return ErrInvalidPasswordFormat
	}
	if !utils.IsValidEmail(req.Email) || len(req.Email) > 255 {
		return ErrInvalidEmailFormat
	}

	return nil
}

func IsValidTweetReq(req *CreateTweetRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return ErrRequiredFieldMissing
	}
	if utf8.RuneCountInString(req.Content) > 1000 {
		return ErrInvalidContentFormat
	}
	return nil
}