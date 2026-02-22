package errcode

import "errors"

var (
	ErrInvalidRequestFormat  = errors.New("リクエスト形式が正しくありません")
	ErrInvalidJSON           = errors.New("JSONの構文が正しくありません")
	ErrInvalidIDFormat       = errors.New("IDの形式が正しくありません")
	ErrEditTimeExpired       = errors.New("投稿から10分以上経過したツイートは編集できません")
	ErrForbidden             = errors.New("指定された操作を行う権限がありません")
	ErrRequiredFieldMissing  = errors.New("必要な項目が不足しています")
	ErrInvalidUsernameFormat = errors.New("ユーザーネームの形式が正しくありません(4〜50文字)")
	ErrInvalidEmailFormat    = errors.New("有効なメールアドレスを入力してください(最大255文字)")
	ErrInvalidPasswordFormat = errors.New("パスワードの形式が正しくありません(8〜72文字)")
	ErrInvalidUserID         = errors.New("無効なユーザーIDです")
	ErrInvalidSessionID      = errors.New("無効なセッションIDです")
	ErrInvalidTokenFormat    = errors.New("有効トークンを入力してください(最大255文字)")
	ErrInvalidUrlFormat      = errors.New("Will be written")
	ErrInvalidTweetID        = errors.New("無効なツイートIDです")
	ErrInvalidContentFormat  = errors.New("contentの形式が正しくありません(最大1000文字)")
	ErrAlreadyFollowing      = errors.New("既にこのユーザーをフォローしています")
	ErrCannotFollowSelf      = errors.New("自分自身をフォローすることはできません")
    ErrNotFollowing          = errors.New("このユーザーをフォローしていません")

	ErrValueTooLong = errors.New("入力内容が長すぎます")

	ErrUserNotFound    = errors.New("ユーザーデータが存在しません")
	ErrSessionNotFound = errors.New("セッションが見つかりません")
	ErrTweetNotFound   = errors.New("ツイートが見つかりません")

	ErrSessionExpired   = errors.New("セッションが期限切れです")
	ErrUsernameConflict = errors.New("ユーザーネームは既に使用されています")
	ErrEmailConflict    = errors.New("メールのアドレスは既に使用されています")
	ErrTokenConflict    = errors.New("トークンは既に存在します")

	ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが正しくありません")

	ErrInternal = errors.New("内部サーバーエラーが発生しました")
)