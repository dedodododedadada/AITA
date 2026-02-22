package dto

import (
	"aita/internal/models"
	"time"
)

type FollowRecord struct {
	ID          int64
	FollowerID  int64
	FollowingID int64
	CreatedAt   time.Time
}

type RelationRecord struct {
	Following  	bool
	FollowedBy 	bool
	IsMutual    bool
}

func NewFollowRecord(follow *models.Follow) *FollowRecord {
	if follow == nil {
		return nil
	}

	return &FollowRecord{
		ID: follow.ID,
		FollowerID: follow.FollowerID,
		FollowingID: follow.FollowingID,
		CreatedAt: follow.CreatedAt,
	}

}

func NewRelationRecord(relation *models.RelationShip) *RelationRecord {
	if relation == nil {
		return nil
	}

	return &RelationRecord{
		Following: relation.Following,
		FollowedBy: relation.FollowedBy,
		IsMutual: relation.Following && relation.FollowedBy,
	}
}