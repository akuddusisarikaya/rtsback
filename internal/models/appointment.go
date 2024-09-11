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
	CompanyName   string             `bson:"company_name,omitempty"`
	CompanyID	  string			 `bson:"company_id,omitempty"`
	Services      []string           `bson:"services,omitempty"`
	Date          time.Time          `bson:"date,omitempty"`
	StartTime     time.Time          `bson:"start_time,omitempty"`
	EndTime       time.Time          `bson:"end_time,omitempty"`
	Activate      bool               `bson:"activate"`
	Notes         string             `bson:"notes,omitempty"`
	CreatedAt     time.Time          `bson:"created_at,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at,omitempty"`
}

type AutoAddRequest struct {
	ProviderEmail string   `bson:"providerEmail,omitempty"`
	CompanyName   string   `bson:"companyName,omitempty"`
	Weekdays      []string `bson:"weekdays,omitempty"`   // Örneğin: ["Monday", "Wednesday"]
	ShiftStart    string   `bson:"shiftStart,omitempty"` // Örneğin: "09:00"
	ShiftEnd      string   `bson:"shiftEnd,omitempty"`   // Örneğin: "17:00"
	Period        int      `bson:"period,omitempty"`     // Dakika cinsinden periyot, örneğin: 30
	Activate      bool     `bson:"activate"`
}
