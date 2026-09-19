package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PollOption struct {
	ID    string `bson:"id" json:"id"`
	Text  string `bson:"text" json:"text"`
	Votes int64  `bson:"votes" json:"votes"`
}

type PollStatus string

const (
	PollStatusActive PollStatus = "active"
	PollStatusClosed PollStatus = "closed"
)

type Poll struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Question    string             `bson:"question" json:"question"`
	Options     []PollOption       `bson:"options" json:"options"`
	CreatorID   primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	CreatorName string             `bson:"creator_name,omitempty" json:"creator_name,omitempty"`
	ShareCode   string             `bson:"share_code" json:"share_code"`
	Status      PollStatus         `bson:"status" json:"status"`
	ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	TotalVotes  int64              `bson:"total_votes" json:"total_votes"`
}

type CreatePollRequest struct {
	Question         string   `json:"question" binding:"required,min=5,max=300"`
	Options          []string `json:"options" binding:"required,min=2,max=10"`
	ExpiresInMinutes int      `json:"expires_in_minutes"`
}

type VoteRequest struct {
	OptionID   string `json:"option_id" binding:"required"`
	VoterToken string `json:"voter_token" binding:"required"`
}

type VoteResultItem struct {
	OptionID   string  `json:"option_id"`
	Text       string  `json:"text"`
	Votes      int64   `json:"votes"`
	Percentage float64 `json:"percentage"`
}

type PollResultsResponse struct {
	PollID     string           `json:"poll_id"`
	Question   string           `json:"question"`
	ShareCode  string           `json:"share_code"`
	Status     PollStatus       `json:"status"`
	TotalVotes int64            `json:"total_votes"`
	Options    []VoteResultItem `json:"options"`
	IsExpired  bool             `json:"is_expired"`
	ExpiresAt  *time.Time       `json:"expires_at,omitempty"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type UpdatePollStatusRequest struct {
	Status PollStatus `json:"status" binding:"required,oneof=active closed"`
}
