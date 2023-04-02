package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserId       string
	Email        string
	PasswordHash string
}

func (u *User) BeforeCreate() {
	u.UserId = uuid.NewString()
}
