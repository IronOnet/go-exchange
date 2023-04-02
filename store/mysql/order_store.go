package mysql

import (
	"github.com/irononet/go-exchange/entities"
	"gorm.io/gorm"
	"time"
)


func (s *Store) GetOrderById(orderId int64) (*entities.Order, error){
	var order *entities.Order 
	err := s.db.Where("id=?", orderId).Scan(&order).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	}
	return order, err 
}

func (s *Store) GetOrderByClientUid(orderId int64, clientUid string) (*entities.Order, error){
	var order entities.Order 
	err := s.db.Where("id=?", orderId).Where("client_uid", clientUid).Scan(&order).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	}
	return &order, err 
}

func (s *Store) GetOrderByIdForUpdate(orderId int64) (*entities.Order, error){
	var order entities.Order 
	err := s.db.Raw("SELECT * FROM orders WHERE id=? FOR UPDATE", orderId).Scan(&order).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	}
	return &order, err 
}

func (s *Store) GetOrderByUserId(userId int64, statuses []entities.OrderStatus, side *entities.Side, productId string, beforeId, afterId int64, limit int) ([]*entities.Order, error){
	db := s.db.Where("user_id =?", userId) 

	if len(statuses) != 0{
		db = db.Where("status IN (?)", statuses)
	}
	if len(productId) != 0{
		db  = db.Where("product_id=?", productId)
	}

	if side != nil{
		db = db.Where("side=?", side)
	}

	if beforeId > 0{
		db = db.Where("side=?", side)
	}
	if afterId > 0{
		db = db.Where("id<?", afterId)
	}

	if limit <= 0{
		limit = 100 
	}

	db = db.Order("id DESC").Limit(limit) 

	var orders []*entities.Order 
	err := db.Find(&orders).Error 
	return orders, err 
}

func (s *Store) AddOrder(order *entities.Order) error{
	order.CreatedAt = time.Now() 
	return s.db.Create(order).Error 
}

func (s *Store) UpdateOrder(order *entities.Order) error{
	order.UpdatedAt = time.Now() 
	return s.db.Save(order).Error 
}

func (s *Store) UpdateOrderStatus(orderId int64, oldStatus, newStatus entities.OrderStatus) (bool, error){
	var order *entities.Order 
	err := s.db.Where("id=?", orderId).Where("status=?", oldStatus).Scan(&order).Error 
	if err != nil{
		return false, err
	}
	order.Status = newStatus 
	order.UpdatedAt = time.Now() 
	return s.db.Save(order).RowsAffected > 0, nil 

}