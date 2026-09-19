package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PollRepository struct {
	pollCol     *mongo.Collection
	voteCol     *mongo.Collection
	isAvailable bool
	mu          sync.RWMutex
	memPolls    map[primitive.ObjectID]*models.Poll
	memByShare  map[string]*models.Poll
	memVotes    map[string]*models.Vote // key: pollID:voterToken
}

func NewPollRepository(db *mongo.Database, isAvailable bool) *PollRepository {
	repo := &PollRepository{
		pollCol:     db.Collection("polls"),
		voteCol:     db.Collection("votes"),
		isAvailable: isAvailable,
		memPolls:    make(map[primitive.ObjectID]*models.Poll),
		memByShare:  make(map[string]*models.Poll),
		memVotes:    make(map[string]*models.Vote),
	}
	if isAvailable {
		repo.initIndexes()
	}
	return repo
}

func (r *PollRepository) initIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r.pollCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"share_code": 1},
		Options: options.Index().SetUnique(true),
	})

	r.pollCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"creator_id": 1, "created_at": -1},
	})

	r.voteCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "poll_id", Value: 1}, {Key: "voter_token", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.CreatedAt = time.Now().UTC()
	if poll.ID.IsZero() {
		poll.ID = primitive.NewObjectID()
	}

	if r.isAvailable {
		result, err := r.pollCol.InsertOne(ctx, poll)
		if err != nil {
			return err
		}
		poll.ID = result.InsertedID.(primitive.ObjectID)
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.memPolls[poll.ID] = poll
	r.memByShare[poll.ShareCode] = poll
	return nil
}

func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	if r.isAvailable {
		var poll models.Poll
		err := r.pollCol.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
		if err != nil {
			return nil, err
		}
		return &poll, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if poll, ok := r.memPolls[id]; ok {
		return poll, nil
	}
	return nil, errors.New("poll not found")
}

func (r *PollRepository) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	if r.isAvailable {
		var poll models.Poll
		err := r.pollCol.FindOne(ctx, bson.M{"share_code": shareCode}).Decode(&poll)
		if err != nil {
			return nil, err
		}
		return &poll, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if poll, ok := r.memByShare[shareCode]; ok {
		return poll, nil
	}
	return nil, errors.New("poll not found")
}

func (r *PollRepository) FindByCreatorID(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	if r.isAvailable {
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

	r.mu.RLock()
	defer r.mu.RUnlock()
	var polls []models.Poll
	for _, poll := range r.memPolls {
		if poll.CreatorID == creatorID {
			polls = append(polls, *poll)
		}
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *PollRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID, status models.PollStatus) error {
	if r.isAvailable {
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

	r.mu.Lock()
	defer r.mu.Unlock()
	poll, ok := r.memPolls[id]
	if !ok || poll.CreatorID != creatorID {
		return errors.New("poll not found or unauthorized")
	}
	poll.Status = status
	return nil
}

func (r *PollRepository) Delete(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID) error {
	if r.isAvailable {
		result, err := r.pollCol.DeleteOne(ctx, bson.M{"_id": id, "creator_id": creatorID})
		if err != nil {
			return err
		}
		if result.DeletedCount == 0 {
			return errors.New("poll not found or unauthorized")
		}
		r.voteCol.DeleteMany(ctx, bson.M{"poll_id": id})
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	poll, ok := r.memPolls[id]
	if !ok || poll.CreatorID != creatorID {
		return errors.New("poll not found or unauthorized")
	}
	delete(r.memByShare, poll.ShareCode)
	delete(r.memPolls, id)
	return nil
}

func (r *PollRepository) HasVoterVoted(ctx context.Context, pollID primitive.ObjectID, voterToken string) (bool, error) {
	if r.isAvailable {
		count, err := r.voteCol.CountDocuments(ctx, bson.M{
			"poll_id":     pollID,
			"voter_token": voterToken,
		})
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", pollID.Hex(), voterToken)
	_, voted := r.memVotes[key]
	return voted, nil
}

func (r *PollRepository) RecordVote(ctx context.Context, vote *models.Vote) error {
	vote.VotedAt = time.Now().UTC()
	if vote.ID.IsZero() {
		vote.ID = primitive.NewObjectID()
	}

	if r.isAvailable {
		_, err := r.voteCol.InsertOne(ctx, vote)
		if err != nil {
			return err
		}
		_, err = r.pollCol.UpdateOne(
			ctx,
			bson.M{"_id": vote.PollID, "options.id": vote.OptionID},
			bson.M{
				"$inc": bson.M{
					"total_votes":     1,
					"options.$.votes": 1,
				},
			},
		)
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s", vote.PollID.Hex(), vote.VoterToken)
	r.memVotes[key] = vote

	if poll, ok := r.memPolls[vote.PollID]; ok {
		poll.TotalVotes++
		for i := range poll.Options {
			if poll.Options[i].ID == vote.OptionID {
				poll.Options[i].Votes++
				break
			}
		}
	}
	return nil
}
