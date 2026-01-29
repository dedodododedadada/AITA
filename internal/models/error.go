package models

import "errors"

var (
	ErrUserNotFound = errors.New("ユーザーデータが存在しません")
	ErrDuplicateEntry = errors.New("リソースは既に存在します")
	ErrContentEmpty = errors.New("ツイート内容を空にすることはできません")
)