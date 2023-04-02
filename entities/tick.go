package entities

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Tick struct {
	gorm.Model

	ProductId int64   `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductId"`

	Granularity int64 `json:"granularity"`
	Time        int64 `json:"time"`

	Open      decimal.Decimal `json:"open_price"`
	High      decimal.Decimal `json:"high_price"`
	Low       decimal.Decimal `json:"low_price"`
	Close     decimal.Decimal `json:"close_price"`
	Volume    decimal.Decimal `json:"volume"`
	LogOffset int64
	LogSeq    int64
}
