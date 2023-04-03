package mysql

import (
	"github.com/irononet/go-exchange/entities" 
	"gorm.io/gorm"
)

func (s *Store) GetTicksByProductId(productId string, granularity int64, limit int) ([]*entities.Tick, error){
	db := s.db.Where("product_id=?", productId).Where("granularity=?", granularity).Order("time DESC").Limit(limit) 
	var ticks []*entities.Tick 
	err := db.Find(&ticks).Error 
	return ticks, err 
}

func (s *Store) GetLastTickByProductId(productId string, granularity int64) (*entities.Tick, error){
	var tick entities.Tick 
	err := s.db.Where("product_id=?", productId).Where("granularity=?", granularity).Find(&tick).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	}
	return &tick, err 
}

func (s *Store) AddTicks(ticks []*entities.Tick) error{
	if len(ticks) == 0{
		return nil 
	}
	return s.db.Create(ticks).Error 
}