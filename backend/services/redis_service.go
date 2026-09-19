package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrAlreadyVotedInRedis = errors.New("voter has already cast a vote for this poll")
)

type RedisService struct {
	client      *redis.Client
	isAvailable bool
	mu          sync.RWMutex
	// Fallback in-memory store if Redis server isn't running locally
	memVotes   map[string]map[string]int64
	memVoters  map[string]map[string]bool
	memPubSubs map[string][]chan string
}

func NewRedisService(redisURL, redisPassword string) *RedisService {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// If not a full redis:// url, assume host:port
		opt = &redis.Options{
			Addr:     redisURL,
			Password: redisPassword,
		}
	}

	rdb := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	svc := &RedisService{
		client:     rdb,
		memVotes:   make(map[string]map[string]int64),
		memVoters:  make(map[string]map[string]bool),
		memPubSubs: make(map[string][]chan string),
	}

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Redis not connected at %s (%v). Running with high-performance In-Memory Engine & Pub/Sub simulation.", redisURL, err)
		svc.isAvailable = false
	} else {
		log.Printf("✅ Connected to Redis successfully at %s", redisURL)
		svc.isAvailable = true
	}

	return svc
}

func (s *RedisService) IsAvailable() bool {
	return s.isAvailable
}

func (s *RedisService) votesKey(pollID string) string {
	return fmt.Sprintf("poll:%s:votes", pollID)
}

func (s *RedisService) votersKey(pollID string) string {
	return fmt.Sprintf("poll:%s:voters", pollID)
}

func (s *RedisService) channelKey(pollID string) string {
	return fmt.Sprintf("poll:%s:events", pollID)
}

// InitPollVotes initializes the Redis hash for a poll
func (s *RedisService) InitPollVotes(ctx context.Context, pollID string, optionIDs []string, initialCounts map[string]int64) error {
	if s.isAvailable {
		key := s.votesKey(pollID)
		pipe := s.client.Pipeline()
		for _, optID := range optionIDs {
			val := int64(0)
			if count, ok := initialCounts[optID]; ok {
				val = count
			}
			pipe.HSet(ctx, key, optID, val)
		}
		pipe.Expire(ctx, key, 30*24*time.Hour)
		_, err := pipe.Exec(ctx)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.memVotes[pollID]; !ok {
		s.memVotes[pollID] = make(map[string]int64)
	}
	for _, optID := range optionIDs {
		val := int64(0)
		if count, ok := initialCounts[optID]; ok {
			val = count
		}
		s.memVotes[pollID][optID] = val
	}
	return nil
}

// CheckVoted checks if voterToken is in the Redis set
func (s *RedisService) CheckVoted(ctx context.Context, pollID string, voterToken string) (bool, error) {
	if s.isAvailable {
		votersKey := s.votersKey(pollID)
		return s.client.SIsMember(ctx, votersKey, voterToken).Result()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if voters, ok := s.memVoters[pollID]; ok {
		return voters[voterToken], nil
	}
	return false, nil
}

// IncrementVote atomically records the voter and increments the vote count in Redis
func (s *RedisService) IncrementVote(ctx context.Context, pollID string, optionID string, voterToken string) (int64, error) {
	if s.isAvailable {
		votersKey := s.votersKey(pollID)
		votesKey := s.votesKey(pollID)

		// Check and add voter atomically using SAdd
		added, err := s.client.SAdd(ctx, votersKey, voterToken).Result()
		if err != nil {
			return 0, err
		}
		if added == 0 {
			return 0, ErrAlreadyVotedInRedis
		}

		// Increment hash field
		newCount, err := s.client.HIncrBy(ctx, votesKey, optionID, 1).Result()
		if err != nil {
			// Rollback voter addition if increment failed
			s.client.SRem(ctx, votersKey, voterToken)
			return 0, err
		}

		return newCount, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.memVoters[pollID]; !ok {
		s.memVoters[pollID] = make(map[string]bool)
	}
	if s.memVoters[pollID][voterToken] {
		return 0, ErrAlreadyVotedInRedis
	}
	s.memVoters[pollID][voterToken] = true

	if _, ok := s.memVotes[pollID]; !ok {
		s.memVotes[pollID] = make(map[string]int64)
	}
	s.memVotes[pollID][optionID]++
	return s.memVotes[pollID][optionID], nil
}

// GetVotes retrieves current vote counts from Redis
func (s *RedisService) GetVotes(ctx context.Context, pollID string) (map[string]int64, error) {
	if s.isAvailable {
		key := s.votesKey(pollID)
		raw, err := s.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		counts := make(map[string]int64)
		for k, v := range raw {
			c, _ := strconv.ParseInt(v, 10, 64)
			counts[k] = c
		}
		return counts, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	counts := make(map[string]int64)
	if stored, ok := s.memVotes[pollID]; ok {
		for k, v := range stored {
			counts[k] = v
		}
	}
	return counts, nil
}

// PublishUpdate publishes an updated poll payload to Redis Pub/Sub
func (s *RedisService) PublishUpdate(ctx context.Context, pollID string, payload interface{}) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if s.isAvailable {
		channel := s.channelKey(pollID)
		return s.client.Publish(ctx, channel, string(bytes)).Err()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	chans := s.memPubSubs[pollID]
	for _, ch := range chans {
		select {
		case ch <- string(bytes):
		default:
		}
	}
	return nil
}

// Subscribe listens to Redis Pub/Sub channel
func (s *RedisService) Subscribe(ctx context.Context, pollID string) (<-chan string, func(), error) {
	if s.isAvailable {
		channel := s.channelKey(pollID)
		pubsub := s.client.Subscribe(ctx, channel)
		out := make(chan string, 50)

		go func() {
			defer close(out)
			for {
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					return
				}
				out <- msg.Payload
			}
		}()

		cleanup := func() {
			pubsub.Close()
		}
		return out, cleanup, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(chan string, 50)
	s.memPubSubs[pollID] = append(s.memPubSubs[pollID], out)

	cleanup := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		subList := s.memPubSubs[pollID]
		for i, ch := range subList {
			if ch == out {
				s.memPubSubs[pollID] = append(subList[:i], subList[i+1:]...)
				close(out)
				break
			}
		}
	}

	return out, cleanup, nil
}
