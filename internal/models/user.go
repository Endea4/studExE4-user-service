package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RoleUser       = "USER"
	RoleDriver     = "DRIVER"
	RoleAdmin      = "ADMIN"
	RoleSuperAdmin = "SUPERADMIN"

	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
	StatusDeleted  = "DELETED"
	StatusBanned   = "BANNED"

	DriverStatusReady     = "READY"
	DriverStatusNotReady  = "NOT_READY"
	DriverStatusOnJob     = "ON_JOB"
	DriverStatusSuspended = "SUSPENDED"

	GenderMale   = "MALE"
	GenderFemale = "FEMALE"
	GenderSecret = "SECRET"
)

type VehicleInfo struct {
	VehicleType  string `bson:"vehicle_type" json:"vehicle_type"`
	LicensePlate string `bson:"license_plate" json:"license_plate"`
	Brand        string `bson:"brand" json:"brand"`
	Model        string `bson:"model" json:"model"`
	Color        string `bson:"color" json:"color"`
	Year         int    `bson:"year" json:"year"`
}

type PerformanceMetrics struct {
	CompletedOrders  int     `bson:"completed_orders" json:"completed_orders"`
	FailedOrders     int     `bson:"failed_orders" json:"failed_orders"`
	TotalRatings     int     `bson:"total_ratings" json:"total_ratings"`
	WeeklyOrderTarget int    `bson:"weekly_order_target" json:"weekly_order_target"`
	WeeklyOrderTotal  int    `bson:"weekly_order_total" json:"weekly_order_total"`
	WeeklyOrderFailed int    `bson:"weekly_order_failed" json:"weekly_order_failed"`
	SuccessRate      float64 `bson:"success_rate" json:"success_rate"`
	AverageRating    float64 `bson:"average_rating" json:"average_rating"`
	TotalEarnings    float64 `bson:"total_earnings" json:"total_earnings"`
}

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Phone       []string           `bson:"phone" json:"phone"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"-"`
	Name        string             `bson:"name" json:"name"`
	DisplayName string             `bson:"display_name" json:"display_name"`
	Gender      string             `bson:"gender" json:"gender"`
	Roles       []string           `bson:"roles" json:"roles"`
	Status      string             `bson:"status" json:"status"`

	ReputationScore float64 `bson:"reputation_score" json:"reputation_score"`
	TotalCancels    int     `bson:"total_cancels" json:"total_cancels"`

	DriverStatus        string             `bson:"driver_status" json:"driver_status,omitempty"`
	PhoneWAOperational  string             `bson:"phone_wa_operational" json:"phone_wa_operational,omitempty"`
	Religion            string             `bson:"religion" json:"religion,omitempty"`
	IsDocumentVerified  bool               `bson:"is_document_verified" json:"is_document_verified"`
	IsCompletedProfile  bool               `bson:"is_completed_profile" json:"is_completed_profile"`
	VehicleInfo         VehicleInfo        `bson:"vehicle_info" json:"vehicle_info,omitempty"`
	PerformanceMetrics  PerformanceMetrics `bson:"performance_metrics" json:"performance_metrics,omitempty"`
	BranchID            primitive.ObjectID `bson:"branch_id" json:"branch_id,omitempty"`
	SchoolID            primitive.ObjectID `bson:"school_id" json:"school_id,omitempty"`
	NIK                 string             `bson:"nik" json:"nik,omitempty"`
	ResignStatus        string             `bson:"resign_status" json:"resign_status,omitempty"`
	Inventory           []string           `bson:"inventory" json:"inventory,omitempty"`
	JoinedDate          time.Time          `bson:"joined_date" json:"joined_date,omitempty"`

	SavedLocations []SavedLocation `bson:"saved_locations" json:"saved_locations"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type SavedLocation struct {
	Name    string  `bson:"name" json:"name"`
	Address string  `bson:"address,omitempty" json:"address,omitempty"`
	Lat     float64 `bson:"lat" json:"lat"`
	Lng     float64 `bson:"lng" json:"lng"`
}
