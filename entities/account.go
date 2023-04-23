package entities

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	UserId    int64
	Currency  string
	Available decimal.Decimal
	Hold      decimal.Decimal
}
