package models

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/go-playground/validator/v10"
)

var (
    ErrInvalidRequestFormat   = errors.New("リクエスト形式が正しくありません")
    ErrInvalidJSON            = errors.New("JSONの構文が正しくありません")

	ErrRequiredFieldMissing   = errors.New("必要な項目が不足しています")
	ErrInvalidUsernameFormat  = errors.New("ユーザーネームの形式が正しくありません(4〜50文字)")
    ErrInvalidEmailFormat     = errors.New("有効なメールアドレスを入力してください(最大255文字)")
    ErrInvalidPasswordFormat  = errors.New("パスワードの形式が正しくありません(8〜72文字)")
	ErrInvalidUserID          = errors.New("無効なユーザーIDです")
    ErrInvalidTokenFormat     = errors.New("有効トークンを入力してください(最大255文字)")
	ErrInvalidUrlFormat       = errors.New("Will be written")
    ErrInvalidContentFormat   = errors.New("contentの形式が正しくありません(最大1000文字)")
	
    ErrValueTooLong           = errors.New("入力内容が長すぎます")

	ErrUserNotFound           = errors.New("ユーザーデータが存在しません")
	ErrSessionNotFound        = errors.New("セッションが見つかりません")
	
	ErrSessionExpired         = errors.New("セッションが期限切れです")
	ErrUsernameConflict       = errors.New("ユーザーネームは既に使用されています")
	ErrEmailConflict          = errors.New("メールのアドレスは既に使用されています")
	ErrTokenConflict          = errors.New("トークンは既に存在します")
	
	ErrInvalidCredentials     = errors.New("メールアドレスまたはパスワードが正しくありません")
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

func GetStatusCode(err error) int {
    if err == nil {
        return 200 
    }

    if errors.Is(err, ErrInvalidCredentials) || 
       errors.Is(err, ErrSessionExpired) || 
       errors.Is(err, ErrSessionNotFound) {
        return 401
    }

    if errors.Is(err, ErrUserNotFound) {
        return 404
    }

    if errors.Is(err, ErrUsernameConflict) || 
       errors.Is(err, ErrEmailConflict) || 
       errors.Is(err, ErrTokenConflict) {
        return 409
    }

    if errors.Is(err, ErrRequiredFieldMissing) ||
       errors.Is(err, ErrInvalidUsernameFormat) ||
       errors.Is(err, ErrInvalidEmailFormat) ||
       errors.Is(err, ErrInvalidPasswordFormat) ||
       errors.Is(err, ErrInvalidUserID) ||
       errors.Is(err, ErrInvalidTokenFormat) ||
       errors.Is(err, ErrInvalidUrlFormat) ||
       errors.Is(err, ErrInvalidContentFormat ) ||
       errors.Is(err, ErrValueTooLong) ||
       errors.Is(err, ErrInvalidJSON) || 
       errors.Is(err, ErrInvalidRequestFormat) {
        return 400
    }

    return 500
}

func GetBusinessCode(err error) string {
    if err == nil {
        return ""
    }

    switch {
    case errors.Is(err, ErrInvalidJSON):
        return "INVALID_JSON_FORMAT"
    case errors.Is(err, ErrInvalidRequestFormat):
        return "INVALID_REQUEST_FORMAT"
    case errors.Is(err, ErrRequiredFieldMissing):
        return "REQUIRED_FIELD_MISSING"
    case errors.Is(err, ErrInvalidUsernameFormat):
        return "INVALID_USERNAME_FORMAT"
    case errors.Is(err, ErrInvalidEmailFormat):
        return "INVALID_EMAIL_FORMAT"
    case errors.Is(err, ErrInvalidTokenFormat):
        return "INVALID_TOKEN_FORMAT"
    case errors.Is(err, ErrInvalidPasswordFormat):
        return "INVALID_PASSWORD_FORMAT"
    case errors.Is(err, ErrInvalidUserID):
        return "INVALID_USER_ID"
    case errors.Is(err, ErrInvalidUrlFormat):
        return "INVALID_URL_FORMAT"
    case errors.Is(err, ErrInvalidContentFormat) :
        return "INVALID_CONTENT_FORMAT"
    case errors.Is(err, ErrValueTooLong):
        return "VALUE_TOO_LONG"

    case errors.Is(err, ErrInvalidCredentials):
        return "INVALID_CREDENTIALS"
    case errors.Is(err, ErrSessionExpired):
        return "SESSION_EXPIRED"
    case errors.Is(err, ErrSessionNotFound):
        return "SESSION_NOT_FOUND"

    case errors.Is(err, ErrUserNotFound):
        return "USER_NOT_FOUND"

    case errors.Is(err, ErrUsernameConflict):
        return "USERNAME_CONFLICT"
    case errors.Is(err, ErrEmailConflict):
        return "EMAIL_CONFLICT"
    case errors.Is(err, ErrTokenConflict):
        return "TOKEN_CONFLICT"

    default:
        return "INTERNAL_SERVER_ERROR"
    }
}