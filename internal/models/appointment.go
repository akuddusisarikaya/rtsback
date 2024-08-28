package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID primitive.ObjectID `bson:"customer_id,omitempty"`
	ProviderID primitive.ObjectID `bson:"provider_id,omitempty"`
	CompanyID  primitive.ObjectID `bson:"company_id,omitempty"`
	Services   []string           `bson:"services,omitempty"`
	Date       time.Time          `bson:"date,omitempty"`
	Status     string             `bson:"status,omitempty"`
	Notes      string             `bson:"notes,omitempty"`
	CreatedAt  time.Time          `bson:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at,omitempty"`
}

