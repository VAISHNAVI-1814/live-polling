package services

import (
	"context"
	"testing"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRedisVoteIncrementAndPubSub(t *testing.T) {
	ctx := context.Background()
	// Using in-memory fallback engine in RedisService
	redisSvc := NewRedisService("localhost:6379", "")

	pollID := primitive.NewObjectID().Hex()
	options := []string{"opt_go", "opt_python", "opt_js"}

	// 1. Initialize poll in Redis
	err := redisSvc.InitPollVotes(ctx, pollID, options, nil)
	if err != nil {
		t.Fatalf("Failed to init poll votes: %v", err)
	}

	// 2. Test initial counts are 0
	counts, err := redisSvc.GetVotes(ctx, pollID)
	if err != nil {
		t.Fatalf("Failed to get votes: %v", err)
	}
	for _, opt := range options {
		if counts[opt] != 0 {
			t.Errorf("Expected 0 votes for %s, got %d", opt, counts[opt])
		}
	}

	// 3. Test Subscribe to Pub/Sub
	ch, cleanup, err := redisSvc.Subscribe(ctx, pollID)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cleanup()

	// 4. Test atomic increment for opt_go
	voter1 := "voter_client_1"
	newCount, err := redisSvc.IncrementVote(ctx, pollID, "opt_go", voter1)
	if err != nil {
		t.Fatalf("Failed to increment vote: %v", err)
	}
	if newCount != 1 {
		t.Errorf("Expected newCount 1, got %d", newCount)
	}

	// 5. Test duplicate vote prevention for voter1
	_, err = redisSvc.IncrementVote(ctx, pollID, "opt_python", voter1)
	if err == nil {
		t.Fatalf("Expected duplicate vote to be rejected for voter1, but it succeeded")
	}

	// 6. Test second voter voting for opt_go
	voter2 := "voter_client_2"
	newCount2, err := redisSvc.IncrementVote(ctx, pollID, "opt_go", voter2)
	if err != nil {
		t.Fatalf("Failed to increment vote for voter2: %v", err)
	}
	if newCount2 != 2 {
		t.Errorf("Expected newCount 2 for opt_go, got %d", newCount2)
	}

	// 7. Verify Redis GetVotes returns exact tally
	countsAfter, err := redisSvc.GetVotes(ctx, pollID)
	if err != nil {
		t.Fatalf("Failed to get updated votes: %v", err)
	}
	if countsAfter["opt_go"] != 2 {
		t.Errorf("Expected 2 votes for opt_go, got %d", countsAfter["opt_go"])
	}
	if countsAfter["opt_python"] != 0 {
		t.Errorf("Expected 0 votes for opt_python, got %d", countsAfter["opt_python"])
	}

	// 8. Test Pub/Sub event delivery
	testPayload := models.PollResultsResponse{
		PollID:     pollID,
		Question:   "What is your favorite programming language?",
		TotalVotes: 2,
		UpdatedAt:  time.Now(),
	}

	err = redisSvc.PublishUpdate(ctx, pollID, testPayload)
	if err != nil {
		t.Fatalf("Failed to publish update: %v", err)
	}

	select {
	case msg := <-ch:
		if msg == "" {
			t.Errorf("Received empty pub/sub message")
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Timed out waiting for pub/sub message")
	}
}

func TestAuthServiceTokenGeneration(t *testing.T) {
	jwtSecret := "test-secret-key-12345"
	authSvc := NewAuthService(nil, jwtSecret)

	user := &models.User{
		ID:    primitive.NewObjectID(),
		Email: "developer@example.com",
		Name:  "Test Developer",
	}

	token, err := authSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatalf("Generated token is empty")
	}
}
