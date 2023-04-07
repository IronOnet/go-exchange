package mysql

import (
	"github.com/irononet/go-exchange/entities"
	"gorm.io/gorm"
)

func (s *Store) GetLastTradeByProduct(productId string) (*entities.Trade, error) {
	var trade entities.Trade
	err := s.db.Where("product_id=?", productId).Order("id DESC").Limit(1).Find(&trade).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &trade, err
}

func (s *Store) GetTradesByProductId(productId string, limit int) ([]*entities.Trade, error) {
	db := s.db.Where("product_id=?", productId).Order("id DESC").Limit(limit)
	var trades []*entities.Trade
	err := db.Find(&trades).Error
	return trades, err

}

func (s *Store) AddTrades(trades []*entities.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	return s.db.Create(trades).Error
}
