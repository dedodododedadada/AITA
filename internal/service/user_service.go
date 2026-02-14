package service

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"errors"
	"fmt"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
}

type PasswordHasher interface {
	Generate(password string) (string, error)
	Compare(hashedPassword, password string) error
}
type userService struct {
	userStore UserStore
	hasher    PasswordHasher
}

func NewUserService(us UserStore, h PasswordHasher) *userService {
	return &userService{
		userStore: us,
		hasher:    h,
	}
}

func (s *userService) Register(ctx context.Context, username, email,password string) (*models.User, error) {
	if username == "" || email == "" || password == "" {
        return nil, errcode.ErrRequiredFieldMissing
    }

	hash, err := s.hasher.Generate(password)
	if err != nil {
		return nil, fmt.Errorf("パスワードをハッシュ化に失敗しました: %w", err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}

	user, err = s.userStore.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("登録に失敗しました: %w", err)
	}
	return user, nil
}

func (s *userService) existsByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, errcode.ErrRequiredFieldMissing
	}

	user, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errcode.ErrUserNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}

	return user, nil
}


func (s *userService) Login(ctx context.Context, email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errcode.ErrRequiredFieldMissing
	}
	user, err := s.existsByEmail(ctx, email) 
	if err != nil {
		return nil, err
	}
	err = s.hasher.Compare(user.PasswordHash, password)
	if err != nil {
		return nil, errcode.ErrInvalidCredentials
	}
	return user, nil
}

func (s *userService) ToMyPage(ctx context.Context, userID int64) (*models.User, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}
	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}
	return user, nil
}
