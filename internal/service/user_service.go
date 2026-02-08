package service

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"errors"
	"fmt"
	"strings"
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
	hasher PasswordHasher
}

func NewUserService(us UserStore, h PasswordHasher) *userService{
	return &userService{
		userStore: us,
		hasher: h,
	}
}

func (s *userService) Register(ctx context.Context, req *models.SignupRequest) (*models.User, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if err := models.IsValidSignUpReq(req); err != nil {
		return nil, err
	}

	hash, err := s.hasher.Generate(req.Password)
	if err != nil {
		return nil, fmt.Errorf("パスワードをハッシュ化に失敗しました: %w", err)
	}

	user := &models.User{
		Username: 		req.Username,
		Email:    		req.Email,
		PasswordHash:   hash,
	}

	user, err = s.userStore.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("登録に失敗しました: %w", err)
	}
	return user, nil
}


func (s *userService) Login(ctx context.Context, email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, models.ErrRequiredFieldMissing
	}
	
	cleanEmail := strings.TrimSpace(email)
	if !utils.IsValidEmail(cleanEmail) {
		return nil,  models.ErrInvalidEmailFormat
	}
	if len(password) < 3 || len (password) > 72 {
		return nil,  models.ErrInvalidPasswordFormat
	}
    
	user, err := s.userStore.GetByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
            return nil, models.ErrInvalidCredentials
        }
        return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}

	err = s.hasher.Compare(user.PasswordHash, password)
	if err != nil {
		return nil, models.ErrInvalidCredentials
	}
	return user, nil
}

func (s *userService) ToMyPage(ctx context.Context, userID int64) (*models.User, error) {
	if userID <= 0 {
		return nil, models.ErrInvalidUserID
	}
	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}
	return user, nil
}
