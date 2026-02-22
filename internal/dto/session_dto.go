package dto

type AuthContext struct {
    UserID    int64
    SessionID int64
    Role      string
}