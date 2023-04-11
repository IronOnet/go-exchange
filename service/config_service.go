package service

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
)

func GetConfigs() ([]*entities.Config, error) {
	return mysql.SharedStore().GetConfigs()
}
