package entities

import "fmt"

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

func NewSideFromString(s string) (*Side, error) {
	side := Side(s)
	switch side {
	case SideBuy:
	case SideSell:
	default:
		return nil, fmt.Errorf("invalid side : %v", s)
	}
	return &side, nil
}

func (s Side) Opposite() Side {
	if s == SideBuy {
		return SideSell
	}
	return SideBuy
}

func (s Side) String() string {
	return string(s)
}

type OrderType string

func (t OrderType) String() string {
	return string(t)
}

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
	OrderStatusOpen       OrderStatus = "OPEN"
	OrderStatusCancelling OrderStatus = "CANCELLING"
	OrderStatusFilled     OrderStatus = "FILLED"
)

func NewOrderStatusFromString(s string) (*OrderStatus, error) {
	status := OrderStatus(s)
	switch status {
	case OrderStatusNew:
	case OrderStatusOpen:
	case OrderStatusCancelling:
	case OrderStatusCancelled:
	case OrderStatusFilled:
	default:
		return nil, fmt.Errorf("invalid status:%v", s)
	}
	return &status, nil
}

type BillType string

type DoneReason string

type TransactionStatus string

const (
	BillTypeTrade              BillType          = "TRADE"
	DoneReasonFilled           DoneReason        = "FILLED"
	DoneReasonCancelled        DoneReason        = "CANCELLED"
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
)
