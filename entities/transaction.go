package entities

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model

	UserId int64
	User   User `gorm:"foreignKey:UserId"`

	Currency   string            `json:"currency"`
	BlockNum   int               `json:"block_num"`
	ConfirmNum int               `json:"confirm_num"`
	Status     TransactionStatus `json:"transaction_status"`
}
