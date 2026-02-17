package models

import "time"

type Follow struct {
	ID              int64       `db:"id"`
	FollowerID  	int64 		`db:"follower_id"`
	FollowingID 	int64 		`db:"following_id"`
	CreatedAt       time.Time   `db:"created_at"`
}

type RelationShip struct {
	Following  bool `db:"following"`
	FollowedBy bool `db:"followed_by"`
} 