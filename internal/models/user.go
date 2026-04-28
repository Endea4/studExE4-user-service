package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents the base user document in MongoDB
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Phone       string             `bson:"phone" json:"phone"`
	Name        string             `bson:"name" json:"name"`
	DisplayName string             `bson:"display_name" json:"display_name"`
	Gender      string             `bson:"gender" json:"gender"`

	// Embedded Customer Profile (if they are a customer)
	CustomerProfile *Customer `bson:"customer_profile,omitempty" json:"customer_profile,omitempty"`

	// Embedded Saved Locations (e.g., "Home", "Campus")
	SavedLocations []SavedLocation `bson:"saved_locations" json:"saved_locations"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Customer holds ride-hailing metrics specific to the user as a passenger
type Customer struct {
	ReputationScore float64 `bson:"reputation_score" json:"reputation_score"`
	TotalOrders     int     `bson:"total_orders" json:"total_orders"`
	TotalCancels    int     `bson:"total_cancels" json:"total_cancels"`
}

// SavedLocation represents a quick-select destination (Home/Campus)
type SavedLocation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Label     string             `bson:"label" json:"label"` // e.g., "Kost", "Fasilkom"
	Latitude  float64            `bson:"latitude" json:"latitude"`
	Longitude float64            `bson:"longitude" json:"longitude"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
