package entities

import (
	"gorm.io/gorm"
)

type Config struct {
	gorm.Model
	Key   string `json:"key"`
	Value string `json:"value"`
}
