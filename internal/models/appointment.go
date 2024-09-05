package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	CustomerEmail string             `bson:"customer_email,omitempty"`
	ProviderEmail string             `bson:"provider_email,omitempty"`
	CustomerName  string             `bson:"customer_name,omitempty"`
	ProviderName  string             `bson:"provider_name,omitempty"`
	CompanyName   string             `bson:"company_id,omitempty"`
	Services      []string           `bson:"services,omitempty"`
	Date          time.Time          `bson:"date,omitempty"`
	StartTime     time.Time          `bson:"start_time,omitempty"`
	EndTime       time.Time          `bson:"end_time,omitempty"`
	Activate      bool               `bson:"activate,omitempty"`
	Notes         string             `bson:"notes,omitempty"`
	CreatedAt     time.Time          `bson:"created_at,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at,omitempty"`
}
 