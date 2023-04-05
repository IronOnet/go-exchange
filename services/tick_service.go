package services

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
	
)

func GetLastTickByProductId(productId string, granularity int64) (*entities.Tick, error){
	return mysql.SharedStore().GetLastTickByProductId(productId, granularity)
}

func GetTicksByProductId(productId string, granularity int64, limit int) ([]*entities.Tick, error){
	return mysql.SharedStore().GetTicksByProductId(productId, granularity, limit)
}

func AddTicks(ticks []*entities.Tick) error{
	if len(ticks) == 0{
		return nil 
	}
	return mysql.SharedStore().AddTicks(ticks)
}