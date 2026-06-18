package repository

import (
	"context"
	"time"

	"github.com/Endea4/studExE4-user-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LeaveRequestRepository struct {
	collection *mongo.Collection
}

func NewLeaveRequestRepository(db *mongo.Database) *LeaveRequestRepository {
	return &LeaveRequestRepository{collection: db.Collection("leave_requests")}
}

func (r *LeaveRequestRepository) Create(ctx context.Context, lr *models.LeaveRequest) error {
	lr.CreatedAt = time.Now()
	lr.Status = "pending"
	_, err := r.collection.InsertOne(ctx, lr)
	return err
}

func (r *LeaveRequestRepository) GetByDriverPhone(ctx context.Context, phone string) ([]models.LeaveRequest, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"driver_phone": phone}, opts)
	if err != nil {
		return nil, err
	}
	var results []models.LeaveRequest
	cursor.All(ctx, &results)
	return results, cursor.Close(ctx)
}

func (r *LeaveRequestRepository) GetByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var lr models.LeaveRequest
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&lr)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

func (r *LeaveRequestRepository) UpdateStatus(ctx context.Context, id, status, reviewedBy string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	now := time.Now()
	update := bson.M{"status": status, "reviewed_by": reviewedBy, "reviewed_at": now}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (r *LeaveRequestRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "driver_phone", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	})
	return err
}
