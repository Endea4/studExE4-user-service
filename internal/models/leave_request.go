package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveRequest struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DriverPhone string            `bson:"driver_phone" json:"driver_phone"`
	Type        string            `bson:"type" json:"type"`
	Reason      string            `bson:"reason" json:"reason"`
	Status      string            `bson:"status" json:"status"`
	ReviewedBy  string            `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time        `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
}
