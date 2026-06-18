package repository

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func timeNow() time.Time {
	return time.Now()
}

func parseObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}
