package models

type Token struct {
	any              `collection:"tokens"`
	BaseModel[Token] `bson:",inline"`
	Name             string `json:"name" bson:"name"`
	Token            string `json:"token" bson:"token"`
}
