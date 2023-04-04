package services

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
	
)

func GetProductById(id string) (*entities.Product, error){
	return mysql.SharedStore().GetProductById(id)
}

func GetProducts() ([]*entities.Product, error){
	return mysql.SharedStore().GetProducts()
}
