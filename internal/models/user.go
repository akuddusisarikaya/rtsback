package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name,omitempty"`
	Email        string             `bson:"email,omitempty"`
	PasswordHash string             `bson:"password_hash,omitempty"`
	Role         string             `bson:"role,omitempty"`
	Phone        string             `bson:"phone,omitempty"`
	CompanyID    primitive.ObjectID `bson:"company_id,omitempty"`
	CreatedAt    time.Time          `bson:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty"`
	SuperUser    bool               `bson:"super_user,omitempty"`
}
