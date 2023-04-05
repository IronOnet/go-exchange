package services

import (
	"errors"
	"fmt"

	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store"
	"github.com/irononet/go-exchange/store/mysql"
	"github.com/shopspring/decimal"
)

func ExecuteBill(userId int64, currency string) error {
	tx, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()

	account, err := tx.GetAccountForUpdate(userId, currency)
	if err != nil {
		return err
	}

	bills, err := tx.GetUnsettledBillsByUserId(userId, currency)

	if err != nil {
		return err
	}

	if len(bills) == 0 {
		return nil
	}

	for _, bill := range bills {
		account.Available = account.Available.Add(bill.Avaiable)
		account.Hold = account.Hold.Add(bill.Hold)

		bill.Settled = true

		err = tx.UpdateBill(bill)
		if err != nil {
			return err
		}
	}

	err = tx.UpdateAccount(account)
	if err != nil {
		return err
	}

	err = tx.CommitTx()
	if err != nil {
		return err
	}

	return nil
}

func HoldBalance(db store.Store, userId int64, currency string, size decimal.Decimal, billType entities.BillType) error {
	if size.LessThanOrEqual(decimal.Zero) {
		return errors.New("size less than 0")
	}

	enough, err := HasEnoughBalance(userId, currency, size)
	if err != nil {
		return err
	}
	if !enough {
		return errors.New(fmt.Sprintf("no enough %v : request=%v", currency, size))
	}

	account, err := db.GetAccountForUpdate(userId, currency)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("no enough")
	}

	account.Available = account.Available.Sub(size)
	account.Hold = account.Hold.Add(size)

	bill := &entities.Bill{
		UserId:   userId,
		Currency: currency,
		Avaiable: size.Neg(),
		Hold:     size,
		Type:     billType,
		Settled:  true,
		Notes:    "",
	}

	err = db.AddBills([]*entities.Bill{bill})

	if err != nil {
		return err
	}

	err = db.UpdateAccount(account)
	if err != nil {
		return err
	}

	return nil
}

func HasEnoughBalance(userId int64, currency string, size decimal.Decimal) (bool, error) {
	account, err := GetAccount(userId, currency)
	if err != nil {
		return false, err
	}
	if account == nil {
		return false, nil
	}
	return account.Available.GreaterThanOrEqual(size), nil
}

func GetAccount(userId int64, currency string) (*entities.Account, error) {
	return mysql.SharedStore().GetAccount(userId, currency)
}

func GetAccountsByUserId(userId int64) ([]*entities.Account, error) {
	return mysql.SharedStore().GetAccountsByUserId(userId)
}

func AddDelayBill(store store.Store, userId int64, currency string, available, hold decimal.Decimal, billType entities.BillType, notes string) (*entities.Bill, error) {
	bill := &entities.Bill{
		UserId:   userId,
		Currency: currency,
		Avaiable: available,
		Hold:     hold,
		Type:     billType,
		Settled:  false,
		Notes:    notes,
	}

	err := store.AddBills([]*entities.Bill{bill})
	return bill, err
}

func GetUnsettledBills() ([]*entities.Bill, error) {
	return mysql.SharedStore().GetUnsettledBills()
}
