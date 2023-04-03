package entities

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Trade struct {
	gorm.Model
	TradeId int64 `json:"trade_id"`
	OrderId int64 `json:"order_id"`
	Order   Order `gorm:"foreignKey:OrderId"`

	MessageSeq int64

	ProductId int64   `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductId"`

	TakerOrderId int64 `json:"taker_order_id"`
	TakerOrderFk Order `gorm:"foreignKey:TakerOrderId"`

	MakerOrderId int64 `json:"maker_order_id"`
	MakerOrderFk Order `gorm:"foreignKey:MakerOrderId"`

	Price decimal.Decimal `json:"price"`
	Size  decimal.Decimal `json:"size"`

	Side Side `json:"side"`
	Time time.Time

	LogOffset int64
	LogSeq    int64
}
