package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	UserId   int64
	Currency uuid.UUID
	Inactive decimal.Decimal
	Active   decimal.Decimal
}
