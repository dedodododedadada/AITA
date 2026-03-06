package dto

import (
	"aita/internal/models"
	"aita/internal/pkg/app"
	"time"
)

type FollowRecord struct {
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

func (r *FollowRecord) ToFollowResponse() *app.FollowResponse {
	return &app.FollowResponse{
		FollowerID: r.FollowerID,
		FollowingID: r.FollowingID,
		CreatedAt: r.CreatedAt,
	}
}

func (r *RelationRecord) ToRelationResponse(userID, targetID int64) *app.RelationResponse {
	if r == nil {
		return &app.RelationResponse{
			MeID: userID,
			TargetID: targetID,
		}
	}
	
	return &app.RelationResponse{
		MeID: userID,
		TargetID: targetID,
		Following: r.Following,
		FollowedBy: r.FollowedBy,
		IsMutual: r.IsMutual,
	}
}