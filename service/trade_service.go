package service

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
)

func GetTradesByProductId(productId string, count int) ([]*entities.Trade, error) {
	return mysql.SharedStore().GetTradesByProductId(productId, count)
}

func AddTrades(trades []*entities.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	return mysql.SharedStore().AddTrades(trades)
}
