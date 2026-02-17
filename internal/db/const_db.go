package db

const (
	errCodeStringDataRightTruncation = "22001"
	errCodeForeignKeyViolation       = "23503"
	errCodeUniqueViolation           = "23505"
	errCodeCheckViolation            = "23514"
	constraintSessionUserFK          = "sessions_user_id_fkey"
	constraintTokenHashUnique        = "sessions_token_hash_key"
	constraintTweetUserFK            = "tweets_user_id_fkey"
	constraintUsernameK              = "users_username_key"
	constraintUseremailK             = "users_email_key"
	constraintTokenhashK             = "sessions_token_hash_key"
	constraintUniqueFollow           = "unique_follow"
	constraintNoSelfFollow           = "no_self_follow"
)
