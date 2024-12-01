package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `json:"id"    bson:"_id,omitempty"`
	Login        string             `json:"login" bson:"login"`
	PasswordHash string             `json:"-"     bson:"password"`
}

type Session struct {
	ID        primitive.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id"    bson:"user_id"`
	Token     string             `json:"token"      bson:"token"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
}

type RegisterInput struct {
	Token    string `json:"token"`
	Login    string `json:"login"`
	Password string `json:"pswd"`
}

type AuthInput struct {
	Login    string `json:"login"`
	Password string `json:"pswd"`
}
