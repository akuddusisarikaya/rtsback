package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Company struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	AdminID         string             `bson:"admin_id,omitempty"`
	AdminName       string             `bson:"admin_name,omitempty"`
	Name            string             `bson:"name,omitempty"`
	Address         string             `bson:"address,omitempty"`
	Phone           string             `bson:"phone,omitempty"`
	CreatedAt       time.Time          `bson:"created_at,omitempty"`
	UpdatedAt       time.Time          `bson:"updated_at,omitempty"`
	ManagersNumber  int                `bson:"managers_number, omitempty"`
	ProvidersNumber int                `bson:"providers_number, omitempty"`
	Services        []string           `bson:"services,omitempty`
}
