package mysql

import (
	"github.com/irononet/go-exchange/entities"
	"gorm.io/gorm"
	"time"
)


func (s *Store) GetAccount(userId int64, currency string) (*entities.Account, error){
	var account entities.Account 
	err := s.db.Where("user_id = ?", userId).Where("currency = ?", currency).Scan(&account).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	}
	return &account, err 
}

func (s *Store) GetAccountsByUserId(userId int64) ([]*entities.Account, error){
	var accounts []*entities.Account 
	db := s.db.Where("user_id = ?", userId).Find(&accounts) 
	err := db.Error 
	if err != nil{
		return nil, nil 
	}
	return accounts, err 
} 

func (s *Store) GetAccountForUpdate(userId int64, currency string) (*entities.Account, error){
	var account entities.Account 
	err := s.db.Where("user_id = ?", userId).Where("currency = ?", currency).Find(&account).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	} 
	return &account, err 
}

func (s *Store) AddAccount(account *entities.Account) error{
	account.CreatedAt = time.Now() 
	return s.db.Create(account).Error 
}

func (s *Store) UpdateAccount(account *entities.Account) error{
	account.UpdatedAt = time.Now() 
	return s.db.Save(account).Error 
}