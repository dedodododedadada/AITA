package dto

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"encoding/json"
	"errors"
	"io"

	"github.com/go-playground/validator/v10"
)


func NewUserResponse(user *models.User) *UserResponse {
	if user == nil {
        return nil
    }
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func NewTweetResponse(tweet *models.Tweet) TweetResponse {
	return TweetResponse{
		ID: 		tweet.ID,
		UserID:     tweet.UserID,
		Content:    tweet.Content,
		ImageURL:   tweet.ImageURL,
		CreatedAt:  tweet.CreatedAt,
		UpdatedAt:  tweet.UpdatedAt,
		IsEdited:   tweet.IsEdited,
	}
}

func Fail(err error) Response {
	return Response{
		Error: err.Error(),
		Code:  GetBusinessCode(err),
	}
}

func Success(data any) Response {
	return Response{
		Data: data,
		Code: "SUCCESS",
	}
}

func SuccessMsg(msg string) Response {
	return Response{
		Message: msg,
		Code: "SUCCESS",
	}
}

func SuccessWithMeta(data any, meta any) Response {
	return Response{
		Data: data,
		Meta: meta,
		Code: "SUCCESS",
	}
}

func FilterBindError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return errcode.ErrInvalidJSON
	}

	var se *json.SyntaxError
	var ute *json.UnmarshalTypeError
	if errors.As(err, &se) || errors.As(err, &ute) {
		return errcode.ErrInvalidJSON
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return errcode.ErrInvalidRequestFormat
	}

	return errcode.ErrInvalidJSON
}

func GetStatusCode(err error) int {
	if err == nil {
		return 200
	}

	if errors.Is(err, errcode.ErrInvalidCredentials) ||
		errors.Is(err, errcode.ErrSessionExpired) ||
		errors.Is(err, errcode.ErrSessionNotFound) {
		return 401
	}

	if errors.Is(err, errcode.ErrForbidden) {
		return 403
	}

	if errors.Is(err, errcode.ErrUserNotFound) ||
		errors.Is(err, errcode.ErrTweetNotFound) {
		return 404
	}

	if errors.Is(err, errcode.ErrUsernameConflict) ||
		errors.Is(err, errcode.ErrEmailConflict) ||
		errors.Is(err, errcode.ErrTokenConflict) {
		return 409
	}

	if errors.Is(err, errcode.ErrEditTimeExpired) {
		return 422
	}
	if errors.Is(err, errcode.ErrRequiredFieldMissing) ||
		errors.Is(err, errcode.ErrInvalidUsernameFormat) ||
		errors.Is(err, errcode.ErrInvalidEmailFormat) ||
		errors.Is(err, errcode.ErrInvalidPasswordFormat) ||
		errors.Is(err, errcode.ErrInvalidUserID) ||
		errors.Is(err, errcode.ErrInvalidSessionID) ||
		errors.Is(err, errcode.ErrInvalidTokenFormat) ||
		errors.Is(err, errcode.ErrInvalidTweetID) ||
		errors.Is(err, errcode.ErrInvalidUrlFormat) ||
		errors.Is(err, errcode.ErrInvalidContentFormat) ||
		errors.Is(err, errcode.ErrValueTooLong) ||
		errors.Is(err, errcode.ErrInvalidJSON) ||
		errors.Is(err, errcode.ErrInvalidIDFormat ) ||
		errors.Is(err, errcode.ErrInvalidRequestFormat) {
		return 400
	}

	return 500
}

func GetBusinessCode(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, errcode.ErrInvalidJSON):
		return "INVALID_JSON_FORMAT"
	case errors.Is(err, errcode.ErrInvalidIDFormat ):
		return "INVALID_ID_FORMAT"
	case errors.Is(err, errcode.ErrInvalidRequestFormat):
		return "INVALID_REQUEST_FORMAT"
	case errors.Is(err, errcode.ErrRequiredFieldMissing):
		return "REQUIRED_FIELD_MISSING"
	case errors.Is(err, errcode.ErrInvalidUsernameFormat):
		return "INVALID_USERNAME_FORMAT"
	case errors.Is(err, errcode.ErrInvalidEmailFormat):
		return "INVALID_EMAIL_FORMAT"
	case errors.Is(err, errcode.ErrInvalidTokenFormat):
		return "INVALID_TOKEN_FORMAT"
	case errors.Is(err, errcode.ErrInvalidPasswordFormat):
		return "INVALID_PASSWORD_FORMAT"
	case errors.Is(err, errcode.ErrInvalidUserID):
		return "INVALID_USER_ID"
	case errors.Is(err, errcode.ErrInvalidSessionID):
		return "INVALID_SESSION_ID"
	case errors.Is(err, errcode.ErrInvalidTweetID):
		return "INVALID_TWEET_ID"
	case errors.Is(err, errcode.ErrInvalidUrlFormat):
		return "INVALID_URL_FORMAT"
	case errors.Is(err, errcode.ErrInvalidContentFormat):
		return "INVALID_CONTENT_FORMAT"
	case errors.Is(err, errcode.ErrValueTooLong):
		return "VALUE_TOO_LONG"

	case errors.Is(err, errcode.ErrInvalidCredentials):
		return "INVALID_CREDENTIALS"
	case errors.Is(err, errcode.ErrSessionExpired):
		return "SESSION_EXPIRED"
	case errors.Is(err, errcode.ErrSessionNotFound):
		return "SESSION_NOT_FOUND"

	case errors.Is(err, errcode.ErrUserNotFound):
		return "USER_NOT_FOUND"
	case errors.Is(err, errcode.ErrTweetNotFound):
		return "TWEET_NOT_FOUND"
	case errors.Is(err, errcode.ErrUsernameConflict):
		return "USERNAME_CONFLICT"
	case errors.Is(err, errcode.ErrEmailConflict):
		return "EMAIL_CONFLICT"
	case errors.Is(err, errcode.ErrTokenConflict):
		return "TOKEN_CONFLICT"
	case errors.Is(err, errcode.ErrForbidden):
		return "FORBIDDEN_ACCESS"

	case errors.Is(err, errcode.ErrEditTimeExpired):
		return "EDIT_TIME_EXPIRED"

	default:
		return "INTERNAL_SERVER_ERROR"
	}
}