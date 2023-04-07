package matching

import (
	"errors"
	"fmt"

	"math"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/irononet/go-exchange/entities"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
)

const (
	orderIdWindowCap = 10000
)

type OrderBook struct {
	product entities.Product

	// bids & asks depth
	depths map[entities.Side]*depth

	// Stricly continuously increasing transaction ID, used for the primary key
	// of the trade
	tradeSeq int64

	// stricly continiously increasing log SEQ, used to write matching log
	LogSeq int64

	// Deduplication measure to prevent the order from repeatedly being
	// submitted to the order book
	orderIdWindow Window
}

type orderBooSnapShot struct {
	// Orderbook product id
	ProductId string

	// All orders
	Orders []BookOrder

	// Trade seq at snapshot time
	TradeSeq int64

	// Log seq at snapshot time
	LogSeq int64

	// State of the duplication window
	OrderIdWindow Window
}

type priceOrderIdKey struct {
	price   decimal.Decimal
	orderId int64
}

func NewOrderBook(product *entities.Product) *OrderBook {
	asks := &depth{
		queue:  treemap.NewWith(priceOrderIdKeyAscComparator),
		orders: map[int64]*BookOrder{},
	}

	bids := &depth{
		queue:  treemap.NewWith(priceOrderIdKeyDescComparator),
		orders: map[int64]*BookOrder{},
	}

	orderBook := &OrderBook{
		product:       *product,
		depths:        map[entities.Side]*depth{entities.SideBuy: bids, entities.SideSell: asks},
		orderIdWindow: newWindow(0, orderIdWindowCap),
	}

	return orderBook
}

func (o *OrderBook) ApplyOrder(order *entities.Order) (logs []Log) {
	// prevent orders from being submitted repeatedly to the matching enginge
	err := o.orderIdWindow.put(int64(order.ID))
	if err != nil {
		log.Error(err)
		return logs
	}

	takerOrder := newBookOrder(order)

	// If it's a Market-Buy Order, set price to infinite high, and if it's a market sell
	// set price to zero, which ensures that prices will cross
	if takerOrder.Type == entities.MARKET_ORDER {
		if takerOrder.Side == entities.SideBuy {
			takerOrder.Price = decimal.NewFromFloat(math.MaxFloat32)
		} else {
			takerOrder.Price = decimal.Zero
		}
	}

	makerDepth := o.depths[takerOrder.Side.Opposite()]
	for itr := makerDepth.queue.Iterator(); itr.Next(); {
		makerOrder := makerDepth.orders[itr.Value().(int64)]

		// check whether ther is price crossing between the taker and
		// the maker
		if (takerOrder.Side == entities.SideBuy && takerOrder.Price.LessThan(makerOrder.Price)) ||
			(takerOrder.Side == entities.SideSell && takerOrder.Price.GreaterThan(makerOrder.Price)) {
			break
		}

		// trade price
		var price = makerOrder.Price

		// trade size
		var size decimal.Decimal

		if takerOrder.Type == entities.LIMIT_ORDER ||
			(takerOrder.Type == entities.MARKET_ORDER && takerOrder.Side == entities.SideSell) {
			if takerOrder.Size.IsZero() {
				break
			}

			// Take the minium size of taker and maker as trade size
			size = decimal.Min(takerOrder.Size, makerOrder.Size)

			// Adjust the size of taker order
			takerOrder.Size = takerOrder.Size.Sub(size)

		} else if takerOrder.Type == entities.MARKET_ORDER && takerOrder.Side == entities.SideBuy {
			if takerOrder.Funds.IsZero() {
				break
			}

			// Calculate the size of taker at current price
			takerSize := takerOrder.Funds.Div(price).Truncate(o.product.BaseScale)
			if takerSize.IsZero() {
				break
			}

			// Taker the minimum size of the taker and maker as trade
			// size
			size = decimal.Min(takerSize, makerOrder.Size)
			funds := size.Mul(price)

			// Adjust the funds of taker order
			takerOrder.Funds = takerOrder.Funds.Sub(funds)
		} else {
			log.Fatal("unknown order type and side combination")
		}

		// Adjust the size of maker order
		err := makerDepth.decrSize(makerOrder.OrderId, size)
		if err != nil {
			log.Fatal(err)
		}

		// mathed, write a log
		matchLog := newMatchLog(o.nextLogSeq(), int64(o.product.ID), o.nextTradeSeq(), takerOrder, makerOrder, price, size)
		logs = append(logs, matchLog)

		// Maker is filled
		if makerOrder.Size.IsZero() {
			doneLog := newDoneLog(o.nextLogSeq(), int64(o.product.ID), makerOrder, makerOrder.Size, entities.DoneReasonFilled)
			logs = append(logs, doneLog)
		}
	}

	if takerOrder.Type == entities.LIMIT_ORDER && takerOrder.Size.GreaterThan(decimal.Zero) {
		// If taker has an uncompleted size, put taker in orderBook
		o.depths[takerOrder.Side].add(*takerOrder)

		openLog := newOpenLog(o.nextLogSeq(), int64(o.product.ID), takerOrder)
		logs = append(logs, openLog)
	} else {
		var remainingSize = takerOrder.Size
		var reason = entities.DoneReasonFilled

		if takerOrder.Type == entities.MARKET_ORDER {
			takerOrder.Price = decimal.Zero
			if (takerOrder.Side == entities.SideSell && takerOrder.Size.GreaterThan(decimal.Zero)) ||
				(takerOrder.Side == entities.SideBuy && takerOrder.Funds.GreaterThan(decimal.Zero)) {
				reason = entities.DoneReasonCancelled
			}
		}

		doneLog := newDoneLog(o.nextLogSeq(), int64(o.product.ID), takerOrder, remainingSize, reason)
		logs = append(logs, doneLog)
	}
	return logs
}

func (o *OrderBook) CancelOrder(order *entities.Order) (logs []Log) {
	_ = o.orderIdWindow.put(int64(order.ID))

	bookOrder, found := o.depths[order.Side].orders[int64(order.ID)]
	if !found {
		return logs
	}

	remainingSize := bookOrder.Size
	err := o.depths[order.Side].decrSize(int64(order.ID), bookOrder.Size)
	if err != nil {
		panic(err)
	}

	doneLog := newDoneLog(o.nextLogSeq(), int64(o.product.ID), bookOrder, remainingSize, entities.DoneReasonCancelled)
	return append(logs, doneLog)
}

func (o *OrderBook) Snapshot() orderBooSnapShot {
	snapshot := orderBooSnapShot{
		Orders:        make([]BookOrder, len(o.depths[entities.SideSell].orders)+len(o.depths[entities.SideBuy].orders)),
		LogSeq:        o.LogSeq,
		TradeSeq:      o.tradeSeq,
		OrderIdWindow: o.orderIdWindow,
	}

	i := 0
	for _, order := range o.depths[entities.SideSell].orders {
		snapshot.Orders[i] = *order
		i++
	}

	for _, order := range o.depths[entities.SideBuy].orders {
		snapshot.Orders[i] = *order
		i++
	}

	return snapshot
}

func (o *OrderBook) Restore(snapshot *orderBooSnapShot) {
	o.LogSeq = snapshot.LogSeq
	o.tradeSeq = snapshot.TradeSeq
	o.orderIdWindow = snapshot.OrderIdWindow
	if o.orderIdWindow.Cap == 0 {
		o.orderIdWindow = newWindow(0, orderIdWindowCap)
	}

	for _, order := range snapshot.Orders {
		o.depths[order.Side].add(order)
	}
}

func (o *OrderBook) nextLogSeq() int64 {
	o.LogSeq++
	return o.LogSeq
}

func (o *OrderBook) nextTradeSeq() int64 {
	o.tradeSeq++
	return o.tradeSeq
}

type depth struct {

	// All orders
	orders map[int64]*BookOrder

	// price first, time first order queue for each order match
	// PriceOrderIdKey - orderId
	queue *treemap.Map
}

func (d *depth) add(order BookOrder) {
	d.orders[order.OrderId] = &order
	d.queue.Put(&priceOrderIdKey{order.Price, order.OrderId}, order.OrderId)
}

func (d *depth) decrSize(orderId int64, size decimal.Decimal) error {
	order, found := d.orders[orderId]
	if !found {
		return errors.New(fmt.Sprintf("order %v not found on book", orderId))
	}

	if order.Size.LessThan(size) {
		return errors.New(fmt.Sprintf("order %v size %v less than %v", orderId, order.Size, size))
	}

	order.Size = order.Size.Sub(size)
	if order.Size.IsZero() {
		delete(d.orders, orderId)
		d.queue.Remove(&priceOrderIdKey{order.Price, order.OrderId})
	}
	return nil
}

type BookOrder struct {
	OrderId int64
	Size    decimal.Decimal
	Funds   decimal.Decimal
	Price   decimal.Decimal
	Side    entities.Side
	Type    entities.OrderType
}

func newBookOrder(order *entities.Order) *BookOrder {
	return &BookOrder{
		OrderId: int64(order.ID),
		Size:    order.Size,
		Funds:   order.Funds,
		Price:   order.Price,
		Side:    order.Side,
		Type:    order.Type,
	}
}

func priceOrderIdKeyAscComparator(a, b interface{}) int {
	aAsserted := a.(*priceOrderIdKey)
	bAsserted := a.(*priceOrderIdKey)

	x := aAsserted.price.Cmp(bAsserted.price)
	if x != 0 {
		return x
	}
	y := aAsserted.orderId - bAsserted.orderId
	if y == 0 {
		return 0
	} else if y > 0 {
		return 1
	} else {
		return -1
	}
}

func priceOrderIdKeyDescComparator(a, b interface{}) int {
	aAsserted := a.(*priceOrderIdKey)
	bAsserted := b.(*priceOrderIdKey)

	x := aAsserted.price.Cmp(bAsserted.price)
	if x != 0 {
		return -x
	}

	y := aAsserted.orderId - bAsserted.orderId
	if y == 0 {
		return 0
	} else if y > 0 {
		return 1
	} else {
		return -1
	}
}
