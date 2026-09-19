package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"live-polling-backend/models"
	"live-polling-backend/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrPollNotFound   = errors.New("poll not found")
	ErrPollClosed     = errors.New("this poll is closed and no longer accepting votes")
	ErrPollExpired    = errors.New("this poll has expired")
	ErrInvalidOption  = errors.New("invalid option selected")
	ErrDuplicateVote  = errors.New("you have already voted on this poll")
)

type PollService struct {
	pollRepo *repository.PollRepository
	redisSvc *RedisService
}

func NewPollService(pollRepo *repository.PollRepository, redisSvc *RedisService) *PollService {
	return &PollService{
		pollRepo: pollRepo,
		redisSvc: redisSvc,
	}
}

func generateShareCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func (s *PollService) CreatePoll(ctx context.Context, creatorID primitive.ObjectID, creatorName string, req models.CreatePollRequest) (*models.Poll, error) {
	// Validate & sanitize options
	seen := make(map[string]bool)
	var sanitizedOptions []models.PollOption

	for i, opt := range req.Options {
		trimmed := strings.TrimSpace(opt)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if seen[lower] {
			return nil, errors.New("duplicate options are not allowed")
		}
		seen[lower] = true

		sanitizedOptions = append(sanitizedOptions, models.PollOption{
			ID:    primitive.NewObjectID().Hex(),
			Text:  trimmed,
			Votes: 0,
		})
		_ = i
	}

	if len(sanitizedOptions) < 2 {
		return nil, errors.New("a poll must have at least 2 unique options")
	}
	if len(sanitizedOptions) > 10 {
		return nil, errors.New("a poll cannot have more than 10 options")
	}

	var expiresAt *time.Time
	if req.ExpiresInMinutes > 0 {
		exp := time.Now().UTC().Add(time.Duration(req.ExpiresInMinutes) * time.Minute)
		expiresAt = &exp
	}

	shareCode := generateShareCode()

	poll := &models.Poll{
		Question:    strings.TrimSpace(req.Question),
		Options:     sanitizedOptions,
		CreatorID:   creatorID,
		CreatorName: creatorName,
		ShareCode:   shareCode,
		Status:      models.PollStatusActive,
		ExpiresAt:   expiresAt,
		TotalVotes:  0,
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}

	// Initialize Redis cache
	optionIDs := make([]string, len(sanitizedOptions))
	for i, o := range sanitizedOptions {
		optionIDs[i] = o.ID
	}
	s.redisSvc.InitPollVotes(ctx, poll.ID.Hex(), optionIDs, nil)

	return poll, nil
}

func (s *PollService) GetPollByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrPollNotFound
	}
	s.checkAndHandleExpiration(ctx, poll)
	return poll, nil
}

func (s *PollService) GetPollByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByShareCode(ctx, strings.ToUpper(strings.TrimSpace(shareCode)))
	if err != nil {
		return nil, ErrPollNotFound
	}
	s.checkAndHandleExpiration(ctx, poll)
	return poll, nil
}

func (s *PollService) GetUserPolls(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	polls, err := s.pollRepo.FindByCreatorID(ctx, creatorID)
	if err != nil {
		return nil, err
	}
	for i := range polls {
		s.checkAndHandleExpiration(ctx, &polls[i])
	}
	return polls, nil
}

func (s *PollService) checkAndHandleExpiration(ctx context.Context, poll *models.Poll) {
	if poll.ExpiresAt != nil && time.Now().UTC().After(*poll.ExpiresAt) {
		poll.Status = models.PollStatusClosed
	}
}

func (s *PollService) UpdateStatus(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID, status models.PollStatus) error {
	err := s.pollRepo.UpdateStatus(ctx, id, creatorID, status)
	if err != nil {
		return err
	}

	// Broadcast status update via Redis Pub/Sub
	results, err := s.GetResults(ctx, id)
	if err == nil {
		s.redisSvc.PublishUpdate(ctx, id.Hex(), results)
	}
	return nil
}

func (s *PollService) DeletePoll(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID) error {
	return s.pollRepo.Delete(ctx, id, creatorID)
}

func (s *PollService) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterToken string) bool {
	// Fast in-memory / Redis check first
	inRedis, err := s.redisSvc.CheckVoted(ctx, pollID.Hex(), voterToken)
	if err == nil && inRedis {
		return true
	}
	// Fallback to MongoDB
	inMongo, _ := s.pollRepo.HasVoterVoted(ctx, pollID, voterToken)
	return inMongo
}

func (s *PollService) Vote(ctx context.Context, pollID primitive.ObjectID, optionID string, voterToken string, ip string) (*models.PollResultsResponse, error) {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil {
		return nil, ErrPollNotFound
	}

	if poll.Status == models.PollStatusClosed {
		return nil, ErrPollClosed
	}

	if poll.ExpiresAt != nil && time.Now().UTC().After(*poll.ExpiresAt) {
		return nil, ErrPollExpired
	}

	// Verify option exists in poll
	validOption := false
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, ErrInvalidOption
	}

	// Fast duplicate vote check
	if s.HasVoted(ctx, pollID, voterToken) {
		return nil, ErrDuplicateVote
	}

	// Atomic Redis increment & voter registration
	_, err = s.redisSvc.IncrementVote(ctx, pollID.Hex(), optionID, voterToken)
	if err != nil {
		if errors.Is(err, ErrAlreadyVotedInRedis) {
			return nil, ErrDuplicateVote
		}
		return nil, err
	}

	// Persist to MongoDB
	vote := &models.Vote{
		PollID:     pollID,
		OptionID:   optionID,
		VoterToken: voterToken,
		IPAddress:  ip,
	}
	if err := s.pollRepo.RecordVote(ctx, vote); err != nil {
		// Log error, but Redis already has the live count
	}

	// Compute updated results and publish to Redis Pub/Sub
	results, err := s.GetResults(ctx, pollID)
	if err != nil {
		return nil, err
	}

	// Broadcast via Redis Pub/Sub to WebSockets
	s.redisSvc.PublishUpdate(ctx, pollID.Hex(), results)

	return results, nil
}

func (s *PollService) GetResults(ctx context.Context, pollID primitive.ObjectID) (*models.PollResultsResponse, error) {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil {
		return nil, ErrPollNotFound
	}

	s.checkAndHandleExpiration(ctx, poll)

	// Fetch live counts from Redis
	liveCounts, err := s.redisSvc.GetVotes(ctx, pollID.Hex())
	if err != nil || len(liveCounts) == 0 {
		// If not in Redis yet, populate from Mongo
		initialCounts := make(map[string]int64)
		optionIDs := make([]string, len(poll.Options))
		for i, o := range poll.Options {
			optionIDs[i] = o.ID
			initialCounts[o.ID] = o.Votes
		}
		s.redisSvc.InitPollVotes(ctx, pollID.Hex(), optionIDs, initialCounts)
		liveCounts = initialCounts
	}

	var totalVotes int64 = 0
	for _, c := range liveCounts {
		totalVotes += c
	}

	var resultItems []models.VoteResultItem
	for _, opt := range poll.Options {
		votes := liveCounts[opt.ID]
		var pct float64 = 0
		if totalVotes > 0 {
			pct = float64(votes) / float64(totalVotes) * 100
		}
		resultItems = append(resultItems, models.VoteResultItem{
			OptionID:   opt.ID,
			Text:       opt.Text,
			Votes:      votes,
			Percentage: pct,
		})
	}

	isExpired := poll.ExpiresAt != nil && time.Now().UTC().After(*poll.ExpiresAt)

	return &models.PollResultsResponse{
		PollID:     poll.ID.Hex(),
		Question:   poll.Question,
		ShareCode:  poll.ShareCode,
		Status:     poll.Status,
		TotalVotes: totalVotes,
		Options:    resultItems,
		IsExpired:  isExpired,
		ExpiresAt:  poll.ExpiresAt,
		UpdatedAt:  time.Now().UTC(),
	}, nil
}
