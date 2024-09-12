package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Name              string             `bson:"name,omitempty"`
	Email             string             `bson:"email,omitempty"`
	EmailVerification bool               `bson:"email_verification,omitempty"`
	PasswordHash      string             `bson:"password_hash,omitempty"`
	Role              string             `bson:"role,omitempty"`
	Phone             string             `bson:"phone,omitempty"`
	PhoneVerification bool               `bson:"phone_verification,omitempty"`
	CompanyID         string             `bson:"company_id,omitempty"`
	CreatedAt         time.Time          `bson:"created_at,omitempty"`
	UpdatedAt         time.Time          `bson:"updated_at,omitempty"`
	SuperUser         bool               `bson:"super_user,omitempty"`
}
