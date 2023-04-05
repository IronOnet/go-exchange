package store

import "github.com/irononet/go-exchange/entities"

type Store interface {

	BeginTx() (Store, error) 
	Rollback() error 
	CommitTx() error 

	// Account store methods
	GetAccount(userId int64, currency string) (*entities.Account, error) 
	GetAccountsByUserId(userId int64) ([]*entities.Account, error) 
	GetAccountForUpdate(userId int64, currency string) (*entities.Account, error) 
	AddAccount(account *entities.Account) error 
	UpdateAccount(account *entities.Account) error  

	// Bill store methods 
	GetUnsettledBillsByUserId(userId int64, currency string) ([]*entities.Bill, error) 
	GetUnsettledBills()([]*entities.Bill, error) 
	AddBills(bills []*entities.Bill) error 
	UpdateBill(bill *entities.Bill) error

	// Config store methods 
	GetConfigs() ([]*entities.Config, error) 

	// Fill store methods 
	GetLastFillByProductId(productId string) (*entities.Fill, error)
	GetUnsettledFillsByOrderId(orderId int64) ([]*entities.Fill, error)
	UpdateFill(fill *entities.Fill) error 
	AddFills(fills []*entities.Fill) error 

	// Order store methods 
	GetOrderById(orderId int64) (*entities.Order, error) 
	GetOrderByClientUid(orderId int64, clientUid string) (*entities.Order, error) 
	GetOrderByIdForUpdate(orderId int64) (*entities.Order, error) 
	GetOrderByUserId(userId int64, statuses []entities.OrderStatus, side *entities.Side, productId string, beforeId, afterId int64, limit int) ([]*entities.Order, error)
	AddOrder(order *entities.Order) error 
	UpdateOrder(order *entities.Order) error 
	UpdateOrderStatus(orderId int64, oldStatus, newStatus entities.OrderStatus) (bool, error)

	// Product store methods 
	GetProductById(id string) (*entities.Product, error) 
	GetProducts()([]*entities.Product, error)

	// Tick store methods 
	GetTicksByProductId(productId string, granularity int64, limit int) ([]*entities.Tick, error) 
	GetLastTickByProductId(productId string, granularity int64) (*entities.Tick, error) 
	AddTicks(ticks []*entities.Tick) error 

	// Trade store methods 
	GetLastTradeByProduct(productId string) (*entities.Trade, error)
	GetTradesByProductId(productId string, limit int) ([]*entities.Trade, error) 
	AddTrades(trades []*entities.Trade) error 

	// User store methods 
	GetUserByEmail(email string) (*entities.User, error) 
	AddUser(user *entities.User) error 

}
