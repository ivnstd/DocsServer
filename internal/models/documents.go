package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Document struct {
	ID      primitive.ObjectID `json:"id"      bson:"_id,omitempty"`
	UserID  primitive.ObjectID `json:"-"       bson:"user_id"`
	Name    string             `json:"name"    bson:"name"`
	Mime    string             `json:"mime"    bson:"mime"`
	File    bool               `json:"file"    bson:"file"`
	Public  bool               `json:"public"  bson:"public"`
	Created time.Time          `json:"created" bson:"created"`
	Grant   []string           `json:"grant"   bson:"grant"`
	JSON    any                `json:"json"    bson:"json,omitempty"`
}

type DocumentMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}
