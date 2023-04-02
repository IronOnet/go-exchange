package entities

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	ProductId int     `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductId"`

	UserId int  `json:"user_id"`
	User   User `gorm:"foreignKey:UserId"`

	ClientUuid    string
	Size          decimal.Decimal `json:"order_size"`
	Funds         decimal.Decimal `json:"funds"`
	FilledSize    decimal.Decimal `json:"filled_size"`
	ExecutedValue decimal.Decimal `json:"executed_value"`
	Price         decimal.Decimal `json:"price"`
	FillFees      decimal.Decimal `json:"fill_fees"`
	Type          OrderType       `json:"order_type"`
	Side          Side            `json:"side"`
	TimeInForce   string          `json:"time_in_force"`
	Status        OrderStatus     `json:"status"`
	Settled       bool            `json:"settled"`
}
