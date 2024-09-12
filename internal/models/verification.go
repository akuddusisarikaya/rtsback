package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Verification struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       string             `bson:"user_id,omitempty"`
	Email        string             `bson:"email,omitempty"`
	EmailCode    string             `bson:"email_code,omitempty"`
	EmailVer     bool               `bson:"email_ver"`
	EmailVerTime time.Time          `bson:"email_ver_time,omitempty"`
	Phone        string             `bson:"phone,omitempty"`
	PhoneVer	 bool				`bson:"phone_ver"`
	PhoneCode    string             `bson:"phone_code,omitempty"`
	PhoneVerTime time.Time          `bson:"phone_ver_time,omitempty"`
	CreateTime   time.Time          `bson:"create_time,omitempty"`
}
