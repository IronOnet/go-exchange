package mysql

import (
	"github.com/irononet/go-exchange/entities"
	"gorm.io/gorm"
)

func (s *Store) GetProductById(id string) (*entities.Product, error) {
	var product entities.Product
	err := s.db.Where("id=?", id).Find(&product).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &product, err
}

func (s *Store) GetProducts() ([]*entities.Product, error) {
	var products []*entities.Product
	err := s.db.Find(&products).Error
	return products, err
}
