package repository

import (
	"context"
	"errors"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PollRepository struct {
	pollCol *mongo.Collection
	voteCol *mongo.Collection
}

func NewPollRepository(db *mongo.Database) *PollRepository {
	repo := &PollRepository{
		pollCol: db.Collection("polls"),
		voteCol: db.Collection("votes"),
	}
	repo.initIndexes()
	return repo
}

func (r *PollRepository) initIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Unique share code index
	r.pollCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"share_code": 1},
		Options: options.Index().SetUnique(true),
	})

	// Creator polls index
	r.pollCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"creator_id": 1, "created_at": -1},
	})

	// Vote uniqueness index: one vote per voterToken per poll
	r.voteCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "poll_id", Value: 1}, {Key: "voter_token", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.CreatedAt = time.Now().UTC()
	result, err := r.pollCol.InsertOne(ctx, poll)
	if err != nil {
		return err
	}
	poll.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	err := r.pollCol.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	var poll models.Poll
	err := r.pollCol.FindOne(ctx, bson.M{"share_code": shareCode}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) FindByCreatorID(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.pollCol.Find(ctx, bson.M{"creator_id": creatorID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *PollRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID, status models.PollStatus) error {
	result, err := r.pollCol.UpdateOne(
		ctx,
		bson.M{"_id": id, "creator_id": creatorID},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("poll not found or unauthorized")
	}
	return nil
}

func (r *PollRepository) Delete(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID) error {
	result, err := r.pollCol.DeleteOne(ctx, bson.M{"_id": id, "creator_id": creatorID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("poll not found or unauthorized")
	}
	// Also clean up votes
	r.voteCol.DeleteMany(ctx, bson.M{"poll_id": id})
	return nil
}

func (r *PollRepository) HasVoterVoted(ctx context.Context, pollID primitive.ObjectID, voterToken string) (bool, error) {
	count, err := r.voteCol.CountDocuments(ctx, bson.M{
		"poll_id":     pollID,
		"voter_token": voterToken,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PollRepository) RecordVote(ctx context.Context, vote *models.Vote) error {
	vote.VotedAt = time.Now().UTC()
	_, err := r.voteCol.InsertOne(ctx, vote)
	if err != nil {
		return err
	}

	// Increment vote tally in Poll document
	_, err = r.pollCol.UpdateOne(
		ctx,
		bson.M{"_id": vote.PollID, "options.id": vote.OptionID},
		bson.M{
			"$inc": bson.M{
				"total_votes":       1,
				"options.$.votes":   1,
			},
		},
	)
	return err
}
