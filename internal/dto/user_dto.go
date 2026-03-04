package dto

import (
	"aita/internal/models"
	"aita/internal/pkg/app"
	"time"
)


type UserRecord struct {
	ID            	int64        	
	Username      	string       	
	Email         	string       	
	PasswordHash  	string       
	CreatedAt     	time.Time    	
	FollowerCount 	int64        	
	FollowingCount  int64         
}

type UserPageRecord struct {
    ID             int64  
    Username       string 
    FollowerCount  int64  
    FollowingCount int64  
}
type UserSlimRecord struct {
	ID            	int64        	
	Username      	string 
}




type UserProfile struct {
	ID            	int64     `json:"id"`   	
	Username      	string    `json:"username"`  
	Email     		string    `json:"email"` 	       	
	FollowerCount 	int64     `json:"follower_count"`   	
	FollowingCount  int64     `json:"following_count"`  
	CreatedAt 		time.Time `json:"created_at"`
}


type UserPage struct {
	ID            	int64     `json:"id"`   	
	Username      	string    `json:"username"`  	       	
	FollowerCount 	int64     `json:"follower_count"`   	
	FollowingCount  int64     `json:"following_count"`  
}



func (u *UserRecord)ToUserResponse() *app.UserResponse {
	if u == nil {
        return nil
    }

	return &app.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func (ur *UserRecord)ToUserModel() *models.User {
	if ur == nil {
		return nil 
	}

	return &models.User{
		ID: ur.ID,
		Username: ur.Username,
		Email: ur.Email,
		PasswordHash: ur.PasswordHash,
		CreatedAt: ur.CreatedAt,
		FollowerCount: ur.FollowerCount,
		FollowingCount: ur.FollowingCount,
	}
}

func (u *UserRecord)ToUserProfile() *UserProfile {
	if u == nil {
        return nil
    }

	return &UserProfile{
		ID:        		u.ID,
		Username:  		u.Username,
		Email:          u.Email,
		FollowerCount: 	u.FollowerCount,
		FollowingCount: u.FollowingCount,
		CreatedAt:      u.CreatedAt,
	}
}

func NewUserRecord(user *models.User) *UserRecord {
	if user == nil {
		return nil
	}

	return &UserRecord{
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt: user.CreatedAt,
		FollowerCount: user.FollowerCount,
		FollowingCount: user.FollowingCount,
	}
}

func NewUserSlimRecord(info *models.UserInfo) *UserSlimRecord {
	if info == nil {
		return nil
	}
	
	return &UserSlimRecord{
		ID: info.ID,
		Username: info.Username,
	}
}

func NewUserPageRecord(info *models.UserInfo, followersCount, followingsCount int64) *UserPageRecord {
	if info == nil || followersCount < 0 || followingsCount < 0 {
		return nil
	}

	return &UserPageRecord{
		ID: info.ID,
		Username: info.Username,
		FollowerCount: followersCount,
		FollowingCount: followingsCount,
	} 
}