package mysql

import (
	"github.com/irononet/go-exchange/entities"
	"gorm.io/gorm"
)

func (s *Store) GetLastFillByProductId(productId string) (*entities.Fill, error) {
	var fill entities.Fill
	err := s.db.Where("product_id = ?", productId).Order("id DESC").Limit(1).Find(&fill).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &fill, err
}

func (s *Store) GetUnsettledFillsByOrderId(orderId int64) ([]*entities.Fill, error) {
	db := s.db.Where("settled = ?", 0).Where("order_id=?", orderId).Order("id ASC").Limit(100)

	var fills []*entities.Fill
	err := db.Find(&fills).Error
	return fills, err
}

func (s *Store) UpdateFill(fill *entities.Fill) error {
	return s.db.Save(fill).Error
}

func (s *Store) AddFills(fills []*entities.Fill) error {
	if len(fills) == 0 {
		return nil
	}
	return s.db.Create(fills).Error
}
