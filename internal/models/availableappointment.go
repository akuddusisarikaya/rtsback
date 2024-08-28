package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AvailebleAppointment struct{
	ID  primitive.ObjectID  `bson:"id,omitempty"`
	ProviderID primitive.ObjectID `bson:"provider_id,omitempty"`
	StartTime time.Ticker `bson:"start_time,omitempty"`
	EndTime time.Ticker `bson:"end_time,omitempty`
	Active bool `bson:"active,omitempty"`
}