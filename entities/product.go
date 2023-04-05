package entities

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	BaseCurrency   string
	QuoteCurrency  string
	BaseMinSize    decimal.Decimal
	QuoteMaxSize   decimal.Decimal
	BaseMaxSize    decimal.Decimal
	QuoteMinSize   decimal.Decimal
	BaseScale      int32
	QuoteScale     int32
	QuoteIncrement float64
}
