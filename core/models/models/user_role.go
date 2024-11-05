package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole struct {
	any                 `collection:"user_roles"`
	BaseModel[UserRole] `bson:",inline"`
	RoleId              primitive.ObjectID `json:"role_id" bson:"role_id"`
	UserId              primitive.ObjectID `json:"user_id" bson:"user_id"`
}
