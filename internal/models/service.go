package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Service struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ServiceName string             `bson:"service_name,omitempty"`
	ProviderID  string             `bson:"provider_id,omitempty"`
	Price       int                `bson:"price,omitempty"`
	CreatedTime time.Time          `bson:"created_time,omitempty"`
}
