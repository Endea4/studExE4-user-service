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

type DebtRepository struct {
	collection *mongo.Collection
}

func NewDebtRepository(db *mongo.Database) *DebtRepository {
	return &DebtRepository{
		collection: db.Collection("debts"),
	}
}

func (r *DebtRepository) GetByDriverPhone(ctx context.Context, driverPhone string) ([]models.Debt, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"driver_phone": bson.M{"$in": []string{driverPhone}},
		"status":       bson.M{"$ne": "paid"},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var debts []models.Debt
	if err = cursor.All(ctx, &debts); err != nil {
		return nil, err
	}
	return debts, nil
}

func (r *DebtRepository) Create(ctx context.Context, debt *models.Debt) error {
	debt.IsActive = true
	_, err := r.collection.InsertOne(ctx, debt)
	return err
}

func (r *DebtRepository) HasActiveDebts(ctx context.Context, phone string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"driver_phone": phone, "is_active": true})
	return count > 0, err
}

func (r *DebtRepository) MarkPaid(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"status": "paid", "paid_at": time.Now(), "is_active": false}})
	return err
}

func (r *DebtRepository) MarkDisputed(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"status": "disputed"}})
	return err
}
