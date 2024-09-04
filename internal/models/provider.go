package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Provider struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name,omitempty"`
	Email       string             `bson:"email,omitempty"`
	Password    string             `bson:"password_hash,omitempty"`
	Phone       string             `bson:"phone,omitempty"`
	Role        string             `bson:"role,omitempty"`
	Services    []string           `bson:"services,omitempty"`
	CompanyName string             `bson:"company_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at,omitempty"`
	UpdatedAt   time.Time          `bson:"updated_at,omitempty"`
}
