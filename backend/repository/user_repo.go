package repository

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection  *mongo.Collection
	isAvailable bool
	mu          sync.RWMutex
	memUsers    map[string]*models.User // by email
	memByID     map[primitive.ObjectID]*models.User
}

func NewUserRepository(db *mongo.Database, isAvailable bool) *UserRepository {
	repo := &UserRepository{
		collection:  db.Collection("users"),
		isAvailable: isAvailable,
		memUsers:    make(map[string]*models.User),
		memByID:     make(map[primitive.ObjectID]*models.User),
	}
	if isAvailable {
		repo.initIndexes()
	}
	return repo
}

func (r *UserRepository) initIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}
	r.collection.Indexes().CreateOne(ctx, indexModel)
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now().UTC()
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	if r.isAvailable {
		result, err := r.collection.InsertOne(ctx, user)
		if err != nil {
			return err
		}
		user.ID = result.InsertedID.(primitive.ObjectID)
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	emailLower := strings.ToLower(user.Email)
	if _, exists := r.memUsers[emailLower]; exists {
		return errors.New("user with this email already exists")
	}
	r.memUsers[emailLower] = user
	r.memByID[user.ID] = user
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	emailLower := strings.ToLower(strings.TrimSpace(email))
	if r.isAvailable {
		var user models.User
		err := r.collection.FindOne(ctx, bson.M{"email": emailLower}).Decode(&user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, exists := r.memUsers[emailLower]; exists {
		return user, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	if r.isAvailable {
		var user models.User
		err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, exists := r.memByID[id]; exists {
		return user, nil
	}
	return nil, mongo.ErrNoDocuments
}
