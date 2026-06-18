package repository

import (
	"context"

	"github.com/Endea4/studExE4-user-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	res, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}
	return nil
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, phone string, updateData bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"phone": phone}, bson.M{"$set": updateData})
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, phone string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"phone": phone})
	return err
}

func (r *UserRepository) CountByFilter(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetActiveDrivers(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"driver_status": models.DriverStatusReady})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []models.User
	cursor.All(ctx, &users)
	return users, nil
}

func (r *UserRepository) GetAllDrivers(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"roles": models.RoleDriver})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []models.User
	cursor.All(ctx, &users)
	return users, nil
}

func (r *UserRepository) IncrementDriverCompletedOrder(ctx context.Context, phoneOrID string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"$or": []bson.M{
		{"phone": phoneOrID},
		{"_id": phoneOrID},
	}}, bson.M{
		"$inc": bson.M{"performance_metrics.completed_orders": 1, "performance_metrics.total_orders": 1},
		"$set": bson.M{"updated_at": timeNow()},
	})
	return err
}

func (r *UserRepository) IncrementDriverFailedOrder(ctx context.Context, phoneOrID string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"$or": []bson.M{
		{"phone": phoneOrID},
		{"_id": phoneOrID},
	}}, bson.M{
		"$inc": bson.M{"performance_metrics.failed_orders": 1},
		"$set": bson.M{"updated_at": timeNow()},
	})
	return err
}

func (r *UserRepository) IncrementDriverEarnings(ctx context.Context, phoneOrID string, amount float64) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"$or": []bson.M{
		{"phone": phoneOrID},
		{"_id": phoneOrID},
	}}, bson.M{
		"$inc": bson.M{"performance_metrics.total_earnings": amount},
		"$set": bson.M{"updated_at": timeNow()},
	})
	return err
}

func (r *UserRepository) SetDriverStatus(ctx context.Context, phone string, status string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"phone": phone}, bson.M{
		"$set": bson.M{"driver_status": status, "updated_at": timeNow()},
	})
	return err
}

func (r *UserRepository) PromoteDriver(ctx context.Context, phone string, vehicleType string, plateNumber string) error {
	update := bson.M{
		"$addToSet": bson.M{"roles": models.RoleDriver},
		"$set": bson.M{
			"driver_status": models.DriverStatusNotReady,
			"resign_status": "active",
			"joined_date":   timeNow(),
			"updated_at":    timeNow(),
		},
	}
	if vehicleType != "" {
		update["$set"].(bson.M)["vehicle_info.vehicle_type"] = vehicleType
	}
	if plateNumber != "" {
		update["$set"].(bson.M)["vehicle_info.license_plate"] = plateNumber
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"phone": phone}, update)
	return err
}
