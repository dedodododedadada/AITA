package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"context"
	"errors"
	"fmt"
)

type UserRepository interface {
	Create(ctx context.Context, record *dto.UserRecord) ( *dto.UserRecord, error)
	GetByEmail(ctx context.Context, email string) (*dto.UserRecord, error) 
	GetByID(ctx context.Context, userID int64) (*dto.UserRecord, error)
	IncreaseFollower(ctx context.Context, id int64, delta int64) error 
	IncreaseFollowing(ctx context.Context, id int64, delta int64) error 
	Exists(ctx context.Context, id int64) (bool, error)
}

type PasswordHasher interface {
	Generate(password string) (string, error)
	Compare(hashedPassword, password string) error
}
type userService struct {
	userRepository 	UserRepository
	hasher    		PasswordHasher
}

func NewUserService(ur UserRepository, h PasswordHasher) *userService {
	return &userService{
		userRepository: ur,
		hasher:    h,
	}
}

func (s *userService) Register(ctx context.Context, username, email,password string) (*dto.UserRecord, error) {
	if username == "" || email == "" || password == "" {
        return nil, errcode.ErrRequiredFieldMissing
    }

	hash, err := s.hasher.Generate(password)
	if err != nil {
		return nil, fmt.Errorf("パスワードをハッシュ化に失敗しました: %w", err)
	}

	user := &dto.UserRecord{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}

	createdUser, err := s.userRepository.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("登録に失敗しました: %w", err)
	}
	return createdUser, nil
}

func (s *userService) existsByEmail(ctx context.Context, email string) (*dto.UserRecord, error) {
	if email == "" {
		return nil, errcode.ErrRequiredFieldMissing
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errcode.ErrUserNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}

	return user, nil
}


func (s *userService) Login(ctx context.Context, email, password string) (*dto.UserRecord, error) {
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

func (s *userService) ToMyPage(ctx context.Context, userID int64) (*dto.UserRecord, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error{
	if userID <= 0 {
		return errcode.ErrInvalidUserID
	}
	err := s.userRepository.IncreaseFollower(ctx , userID, delta) 
	if err != nil {
		return fmt.Errorf("FollowerCountの更新に失敗しました: %w", err)
	}

	return nil
}

func (s *userService) UpdateFollowingCount(ctx context.Context, userID int64, delta int64) error{
	if userID <= 0 {
		return errcode.ErrInvalidUserID
	}

	err := s.userRepository.IncreaseFollowing(ctx , userID, delta)  
		if err != nil {
		return fmt.Errorf("FollowingCountの更新に失敗しました: %w", err)
	}

	return nil
}

func (s *userService) Exists(ctx context.Context, userID int64) (bool, error) {
	if userID <= 0 {
		return false, errcode.ErrInvalidUserID
	}

	exist, err := s.userRepository.Exists(ctx, userID)

	if err != nil {
		return false, fmt.Errorf("ユーザーの存在の確認に失敗しました: %w", err)
	}

	return exist, err
}