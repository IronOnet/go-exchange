package services

import (
	"crypto/md5"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"errors"

	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
)


func CreateUser(email, password string) (*entities.User, error){
	if len(password) < 6{
		return nil, errors.New("password must be of minimum 6 character length")
	}
	user, err := GetUserByEmail(email) 
	if err != nil{
		return nil, err 
	}
	if user != nil{
		return nil, errors.New("email address is already registered")
	}

	user = &entities.User{
		Email: email, 
		PasswordHash: encryptPassword(password),
	}

	return user, mysql.SharedStore().AddUser(user)
}

func RefreshAccessToken(email, password string) (string, error){
	user, err := GetUserByEmail(email) 
	if err != nil{
		return "", err 
	}
	if user == nil{
		return "", errors.New("email not found or password error")
	}
	if user.PasswordHash != encryptPassword(password){
		return "", errors.New("email not found or password error")
	}

	claim := jwt.MapClaims{
		"id": user.ID, 
		"email": user.Email, 
		"passwordHash": user.PasswordHash, 
		"expiresAt": time.Now().Unix(),
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim) 
	return token.SignedString([]byte(conf.GetConfig().JwtSecret))
}

func CheckToken(tokenStr string) (*entities.User, error){
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error){
		return []byte(conf.GetConfig().JwtSecret), nil 
	})
	if err != nil{
		return nil, err 
	}
	claim, ok := token.Claims.(jwt.MapClaims) 
	if ! ok{
		return nil, errors.New("cannot convert claim to MapClaims")
	}
	if !token.Valid{
		return nil, errors.New("token is invalid")
	}

	emailVal, found := claim["email"] 
	if !found{
		return nil, errors.New("bad token")
	}
	email := emailVal.(string) 

	passwordHashVal, found := claim["passwordHash"] 
	if !found{
		return nil, errors.New("bad token")
	}
	passwordHash := passwordHashVal.(string) 

	user, err := GetUserByEmail(email) 
	if err != nil{
		return nil, err 
	}
	if user == nil{
		return nil, errors.New("bad token")
	}
	if user.PasswordHash != passwordHash{
		return nil, errors.New("bad token")
	}
	return user, nil 
}

func ChangePassword(email, newPassword string) error{
	user, err := GetUserByEmail(email) 
	if err != nil{
		return err 
	}
	if user == nil{
		return errors.New("user not found")
	}
	user.PasswordHash = encryptPassword(newPassword) 
	return mysql.SharedStore().UpdateUser(user)
}



func GetUserByEmail(email string) (*entities.User, error){
	return mysql.SharedStore().GetUserByEmail(email)
}

func GetUserByPassword(email, password string) (*entities.User, error){
	user, err := GetUserByEmail(email) 
	if err != nil{
		return nil, err 
	}
	if user == nil || user.PasswordHash != encryptPassword(password){
		return nil, errors.New("user not found or password incorect")
	}
	return user, nil 
}

func encryptPassword(password string) string{
	hash := md5.Sum([]byte(password)) 
	return fmt.Sprintf("%x", hash)
}