package entities

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Fill struct {
	gorm.Model
	TradeId int64 `json:"trade_id"`
	Trade   Trade `gorm:"foreignKey:TradeId"`

	OrderId int64 `json:"order_id"`
	Order   Order `gorm:"foreignKey:OrderId"`

	ProductId int64   `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductId"`

	Size      decimal.Decimal `json:"fill_size"`
	Price     decimal.Decimal `json:"price"`
	Funds     decimal.Decimal `json:"funds"`
	Fee       decimal.Decimal `json:"fees"`
	Liquidity string          `json:"liquidity"`
	Settled   bool            `json:"settled"`
	Side      Side            `json:"side"`
	Done      bool            `json:"done"`

	DoneReason DoneReason `json:"done_reason"`
	LogOffset  int64
	LogSeq     int64

	MessageSeq int64 `gorm:"index:o_m unique"`
}
