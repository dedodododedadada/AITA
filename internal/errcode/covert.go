package errcode

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func FilterBindError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return ErrInvalidJSON
	}

	var se *json.SyntaxError
	var ute *json.UnmarshalTypeError
	if errors.As(err, &se) || errors.As(err, &ute) {
		return ErrInvalidJSON
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return ErrInvalidRequestFormat
	}

	return ErrInvalidJSON
}

type errorMetadata struct {
	HTTPStatus   int
	BusinessCode string
}


var errorToMetadata = map[error]errorMetadata{
	// 400 Bad Request
	ErrInvalidRequestFormat:  {http.StatusBadRequest, "INVALID_REQUEST_FORMAT"},
	ErrInvalidJSON:           {http.StatusBadRequest, "INVALID_JSON_FORMAT"},
	ErrInvalidIDFormat:       {http.StatusBadRequest, "INVALID_ID_FORMAT"},
	ErrRequiredFieldMissing:  {http.StatusBadRequest, "REQUIRED_FIELD_MISSING"},
	ErrInvalidUsernameFormat: {http.StatusBadRequest, "INVALID_USERNAME_FORMAT"},
	ErrInvalidEmailFormat:    {http.StatusBadRequest, "INVALID_EMAIL_FORMAT"},
	ErrInvalidPasswordFormat: {http.StatusBadRequest, "INVALID_PASSWORD_FORMAT"},
	ErrInvalidUserID:         {http.StatusBadRequest, "INVALID_USER_ID"},
	ErrInvalidSessionID:      {http.StatusBadRequest, "INVALID_SESSION_ID"},
	ErrInvalidTokenFormat:    {http.StatusBadRequest, "INVALID_TOKEN_FORMAT"},
	ErrInvalidTweetID:        {http.StatusBadRequest, "INVALID_TWEET_ID"},
	ErrInvalidUrlFormat:      {http.StatusBadRequest, "INVALID_URL_FORMAT"},
	ErrInvalidContentFormat:  {http.StatusBadRequest, "INVALID_CONTENT_FORMAT"},
	ErrValueTooLong:          {http.StatusBadRequest, "VALUE_TOO_LONG"},
	ErrAlreadyFollowing:      {http.StatusBadRequest, "ALREADY_FOLLOWING"}, 
	ErrCannotFollowSelf:      {http.StatusBadRequest, "CANNOT_FOLLOW_SELF"}, 
	ErrNotFollowing:          {http.StatusBadRequest, "NOT_FOLLOWING"},    

	// 401 Unauthorized
	ErrInvalidCredentials: {http.StatusUnauthorized, "INVALID_CREDENTIALS"},
	ErrSessionExpired:     {http.StatusUnauthorized, "SESSION_EXPIRED"},
	ErrSessionNotFound:    {http.StatusUnauthorized, "SESSION_NOT_FOUND"},

	// 403 Forbidden
	ErrForbidden: {http.StatusForbidden, "FORBIDDEN_ACCESS"},

	// 404 Not Found
	ErrUserNotFound:  {http.StatusNotFound, "USER_NOT_FOUND"},
	ErrTweetNotFound: {http.StatusNotFound, "TWEET_NOT_FOUND"},

	// 409 Conflict
	ErrUsernameConflict: {http.StatusConflict, "USERNAME_CONFLICT"},
	ErrEmailConflict:    {http.StatusConflict, "EMAIL_CONFLICT"},
	ErrTokenConflict:    {http.StatusConflict, "TOKEN_CONFLICT"},

	// 422 Unprocessable Entity
	ErrEditTimeExpired: {http.StatusUnprocessableEntity, "EDIT_TIME_EXPIRED"},
}

func GetStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	for targetErr, meta := range errorToMetadata {
		if errors.Is(err, targetErr) {
			return meta.HTTPStatus
		}
	}

	return http.StatusInternalServerError
}

func GetBusinessCode(err error) string {
	if err == nil {
		return ""
	}

	for targetErr, meta := range errorToMetadata {
		if errors.Is(err, targetErr) {
			return meta.BusinessCode
		}
	}

	return "INTERNAL_SERVER_ERROR"
}
