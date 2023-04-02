package entities

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Bill struct {
	gorm.Model
	UserId   int64
	User     User `gorm:"foreignKey:UserId"`
	Currency string
	Avaiable decimal.Decimal
	Hold     decimal.Decimal
	Type     BillType
	Settled  bool
	Notes    string
}
