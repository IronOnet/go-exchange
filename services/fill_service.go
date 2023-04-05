package services

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
)

func GetUnsettledFills(count int) ([]*entities.Fill, error) {
	return mysql.SharedStore().GetUnsettledFills(count)
}

func AddFills(fills []*entities.Fill) error {
	if len(fills) == 0 {
		return nil
	}

	err := mysql.SharedStore().AddFills(fills)
	if err != nil {
		return err
	}

	return nil
}
