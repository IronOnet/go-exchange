package mysql 

import (
	"github.com/irononet/go-exchange/entities" 
	"time" 
	
)

func (s *Store) GetUnsettledBillsByUserId(userId int64, currency string) ([]*entities.Bill, error){
	var bills []*entities.Bill 
	db := s.db.Where("user_id = ?", userId).Where("currency = ?", currency).Order("id ASC").Limit(100) 
	err := db.Find(&bills).Error 
	return bills, err 
}

func (s *Store) GetUnsettledBills() ([]*entities.Bill, error){
	db := s.db.Where("settled = ?", 0).Order("id ASC").Limit(100) 

	var bills []*entities.Bill
	err := db.Find(&bills).Error 
	return bills, err 
}

func (s *Store) AddBills(bills []*entities.Bill) error{
	return s.db.Create(bills).Error 
}

func (s *Store) UpdateBill(bill *entities.Bill) error{
	bill.UpdatedAt = time.Now() 
	return s.db.Save(bill).Error 
}
