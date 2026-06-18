package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Debt struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DriverPhone    string             `bson:"driver_phone" json:"driver_phone"`
	OrderID        primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderNumber    string             `bson:"order_number" json:"order_number"`
	Amount         float64            `bson:"amount" json:"amount"`
	Remaining      float64            `bson:"remaining" json:"remaining"`
	Status         string             `bson:"status" json:"status"`
	IsActive       bool               `bson:"is_active" json:"is_active"`
	Description    string             `bson:"description" json:"description"`
	DueDate        *time.Time         `bson:"due_date,omitempty" json:"due_date,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	PaidAt         *time.Time         `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
}
