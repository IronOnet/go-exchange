package mysql

import (
	"time"

	"github.com/irononet/go-exchange/entities"
	"gorm.io/gorm"
)

func (s *Store) GetUserByEmail(email string) (*entities.User, error){
	var user entities.User 
	err := s.db.Where("email=?", email).Find(&user).Error 
	if err == gorm.ErrRecordNotFound{
		return nil, nil 
	}
	return &user, err 
}

func (s *Store) AddUser(user *entities.User) error{
	user.CreatedAt = time.Now() 
	return s.db.Create(user).Error 
}

func (s *Store) UpdateUser(user *entities.User) error{
	return s.db.Save(user).Error 
}