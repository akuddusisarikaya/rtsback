package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID string             `bson:"customer_id,omitempty"`
	ProviderID string             `bson:"provider_id,omitempty"`
	CompanyID  string             `bson:"company_id,omitempty"`
	Services   []string           `bson:"services,omitempty"`
	Date       time.Time          `bson:"date,omitempty"`
	Status     string             `bson:"status,omitempty"`
	Notes      string             `bson:"notes,omitempty"`
	CreatedAt  time.Time          `bson:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at,omitempty"`
}
