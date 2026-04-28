package repository

import (
	"context"

	"github.com/Endea4/studExE4-user-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// CreateUser inserts a new user document (used when they first message the bot)
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// GetUserByPhone fetches a user by phone number
func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	filter := bson.M{"phone": phone}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user's details (used during the personalization flow)
func (r *UserRepository) UpdateUser(ctx context.Context, phone string, updateData bson.M) error {
	filter := bson.M{"phone": phone}
	update := bson.M{"$set": updateData}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteUser removes a user by their phone number
func (r *UserRepository) DeleteUser(ctx context.Context, phone string) error {
	filter := bson.M{"phone": phone}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// GetAllUsers is just for testing purposes to see all documents
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
