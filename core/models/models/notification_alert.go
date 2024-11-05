package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type NotificationAlert struct {
	any                          `collection:"notification_alerts"`
	BaseModel[NotificationAlert] `bson:",inline"`
	Name                         string             `json:"name" bson:"name"`
	Description                  string             `json:"description" bson:"description"`
	Enabled                      bool               `json:"enabled" bson:"enabled"`
	HasMetricTarget              bool               `json:"has_metric_target" bson:"has_metric_target"`
	MetricTargetId               primitive.ObjectID `json:"metric_target_id,omitempty" bson:"metric_target_id,omitempty"`
	MetricName                   string             `json:"metric_name" bson:"metric_name"`
	Operator                     string             `json:"operator" bson:"operator"`
	LastingSeconds               int                `json:"lasting_seconds" bson:"lasting_seconds"`
	TargetValue                  float32            `json:"target_value" bson:"target_value"`
	Level                        string             `json:"level" bson:"level"`
	TemplateKey                  string             `json:"template_key,omitempty" bson:"template_key,omitempty"`
}