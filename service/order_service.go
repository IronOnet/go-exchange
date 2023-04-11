package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store/mysql"
	"github.com/shopspring/decimal"
)

func PlaceOrder(userId int64, clientUid string, productId string, orderType entities.OrderType,
	side entities.Side, size, price, funds decimal.Decimal) (*entities.Order, error) {

	product, err := GetProductById(productId)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, fmt.Errorf("size %v less than base min size %v", size, product.BaseMinSize)
	}

	if orderType == entities.LIMIT_ORDER {
		size = size.Round(product.BaseScale)
		if size.LessThan(product.BaseMinSize) {
			return nil, fmt.Errorf("size %v less than base min size %v", size, product.BaseMinSize)
		}
		price = price.Round(product.QuoteScale)
		if price.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("price %v less than 0", price)
		}
		funds = size.Mul(price)
	} else if orderType == entities.MARKET_ORDER {
		if side == entities.SideBuy {
			size = decimal.Zero
			price = decimal.Zero
			funds = funds.Round(product.QuoteScale)
			if funds.LessThan(product.QuoteMinSize) {
				return nil, fmt.Errorf("funds %v less than quote min size %v", funds, product.QuoteMinSize)
			}
		} else {
			size = size.Round(product.BaseScale)
			if size.LessThan(product.BaseMinSize) {
				return nil, fmt.Errorf("size %v less than base min size %v", size, product.BaseMinSize)
			}
			price = decimal.Zero
			funds = decimal.Zero
		}
	} else {
		return nil, errors.New("unknown order type")
	}

	var holdCurrency string
	var holdSize decimal.Decimal
	if side == entities.SideBuy {
		holdCurrency, holdSize = product.QuoteCurrency, funds
	} else {
		holdCurrency, holdSize = product.BaseCurrency, size
	}

	order := &entities.Order{
		ClientUuid: clientUid,
		UserId:     int(userId),
		ProductId:  int(product.ID),
		Side:       side,
		Size:       size,
		Funds:      funds,
		Price:      price,
		Status:     entities.OrderStatusNew,
		Type:       orderType,
	}

	// transaction
	db, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Rollback() }()

	err = HoldBalance(db, userId, holdCurrency, holdSize, entities.BillTypeTrade)
	if err != nil {
		return nil, err
	}

	err = db.AddOrder(order)
	if err != nil {
		return nil, err
	}

	return order, db.CommitTx()
}

func UpdateOrderStatus(orderId int64, oldStatus, newStatus entities.OrderStatus) (bool, error) {
	return mysql.SharedStore().UpdateOrderStatus(orderId, oldStatus, newStatus)
}

func ExecuteFill(orderId int64) error {
	// tx
	db, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = db.Rollback() }()

	order, err := db.GetOrderByIdForUpdate(orderId)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found: %v", orderId)
	}
	if order.Status == entities.OrderStatusFilled || order.Status == entities.OrderStatusCancelled {
		return fmt.Errorf("order status invalid: %v %v", orderId, order.Status)
	}

	product, err := GetProductById(strconv.Itoa(order.ProductId))
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found: %v", order.ProductId)
	}
	fills, err := mysql.SharedStore().GetUnsettledFillsByOrderId(orderId)
	if err != nil {
		return err
	}
	if len(fills) == 0 {
		return nil
	}

	var bills []*entities.Bill
	for _, fill := range fills {
		fill.Settled = true
		notes := fmt.Sprintf("%v-%v", fill.OrderId, fill.ID)

		if !fill.Done {
			executedValue := fill.Size.Mul(fill.Price)
			order.ExecutedValue = order.ExecutedValue.Add(executedValue)
			order.FilledSize = order.FilledSize.Add(fill.Size)

			if order.Side == entities.SideBuy {
				bill, err := AddDelayBill(db, int64(order.UserId), product.BaseCurrency, fill.Size, decimal.Zero, entities.BillTypeTrade, notes)
				if err != nil {
					return err
				}
				bills = append(bills, bill)
			} else {
				bill, err := AddDelayBill(db, int64(order.UserId), product.QuoteCurrency, executedValue, decimal.Zero, entities.BillTypeTrade, notes)
				if err != nil {
					return err
				}
				bills = append(bills, bill)
			}
		} else {
			if fill.DoneReason == entities.DoneReasonCancelled {
				order.Status = entities.OrderStatusCancelled
			} else if fill.DoneReason == entities.DoneReasonFilled {
				order.Status = entities.OrderStatusFilled
			} else {
				log.Fatalf("unkown done reason: %v", fill.DoneReason)
			}

			if order.Side == entities.SideBuy {
				remainingFunds := order.Funds.Sub(order.ExecutedValue)
				if remainingFunds.GreaterThan(decimal.Zero) {
					bill, err := AddDelayBill(db, int64(order.UserId), product.QuoteCurrency, remainingFunds, remainingFunds.Neg(), entities.BillTypeTrade, notes)
					if err != nil {
						return err
					}
					bills = append(bills, bill)
				}
			} else {
				remainingSize := order.Size.Sub(order.FilledSize)
				if remainingSize.GreaterThan(decimal.Zero) {
					bill, err := AddDelayBill(db, int64(order.UserId), product.BaseCurrency, remainingSize, remainingSize.Neg(), entities.BillTypeTrade, notes)
					if err != nil {
						return err
					}
					bills = append(bills, bill)
				}
			}
			break
		}
	}

	err = db.UpdateOrder(order)
	if err != nil {
		return err
	}

	for _, fill := range fills {
		err = db.UpdateFill(fill)
		if err != nil {
			return err
		}
	}
	return db.CommitTx()
}

func GetOrderById(orderId int64) (*entities.Order, error) {
	return mysql.SharedStore().GetOrderById(orderId)
}

func GetOrderByClientUid(userId int64, clientUuid string) (*entities.Order, error) {
	return mysql.SharedStore().GetOrderByClientUid(userId, clientUuid)
}

func GetOrderByUserId(userId int64, statuses []entities.OrderStatus, side *entities.Side, productId string,
	beforeId, afterId int64, limit int) ([]*entities.Order, error) {
	return mysql.SharedStore().GetOrderByUserId(userId, statuses, side, productId, beforeId, afterId, limit)
}
