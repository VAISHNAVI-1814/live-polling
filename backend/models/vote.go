package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vote struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID     primitive.ObjectID `bson:"poll_id" json:"poll_id"`
	OptionID   string             `bson:"option_id" json:"option_id"`
	VoterToken string             `bson:"voter_token" json:"voter_token"`
	IPAddress  string             `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	VotedAt    time.Time          `bson:"voted_at" json:"voted_at"`
}
