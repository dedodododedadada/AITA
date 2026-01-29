package db

import "errors"

var (
	ErrConflict= errors.New("レコードが既に存在します")
	ErrNotFound = errors.New("データが存在しません")
	ErrExpired = errors.New("期限が切れているので、無効です")
)
// will be removed in service layer