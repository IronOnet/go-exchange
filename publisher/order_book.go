package publisher

import (
	"encoding/json"
	"fmt"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/go-redis/redis"
	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/utils"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

const (
	ORDER_BOOK_L2_SNAPSHOT_KEY_PREFIX   = "order_book_level2_snapshot_"
	ORDER_BOOK_FULL_SNAPSHOT_KEY_PREFIX = "order_book_full_snapshot_"
)

type OrderBook struct {
	ProductId string
	Seq       int64
	LogOffset int64
	LogSeq    int64
	Depths    map[entities.Side]*treemap.Map
	Orders    map[int64]*matching.BookOrder
}

type OrderBookLevel2Snapshot struct {
	ProductId string
	Seq       int64
	Asks      [][3]interface{}
	Bids      [][3]interface{}
}

type OrderBookFullSnapshot struct {
	ProductId string
	Seq       int64
	LogOffset int64
	LogSeq    int64
	Orders    []matching.BookOrder
}

type PriceLevel struct {
	Price      decimal.Decimal
	Size       decimal.Decimal
	OrderCount int64
}

func NewOrderBook(productId string) *OrderBook {
	b := &OrderBook{
		ProductId: productId,
		Depths:    map[entities.Side]*treemap.Map{},
		Orders:    map[int64]*matching.BookOrder{},
	}

	b.Depths[entities.SideBuy] = treemap.NewWith(utils.DecimalDescComparator)
	b.Depths[entities.SideSell] = treemap.NewWith(utils.DecimalAscComparator)
	return b
}

func (s *OrderBook) SaveOrder(logOffset, logSeq int64, orderId int64, newSize, price decimal.Decimal, side entities.Side) *Level2Change {
	if newSize.LessThan(decimal.Zero) {
		panic(newSize)
	}

	var changedLevel *PriceLevel

	priceLevels := s.Depths[side]
	order, found := s.Orders[orderId]
	if !found {
		if newSize.IsZero() {
			return nil
		}

		s.Orders[orderId] = &matching.BookOrder{
			OrderId: orderId,
			Size:    newSize,
			Side:    side,
			Price:   price,
		}

		val, found := priceLevels.Get(price)
		if !found {
			changedLevel = &PriceLevel{
				Price:      price,
				Size:       newSize,
				OrderCount: 1,
			}
			priceLevels.Put(price, changedLevel)
		} else {
			changedLevel = val.(*PriceLevel)
			changedLevel.Size = changedLevel.Size.Add(newSize)
			changedLevel.OrderCount++
		}
	} else {
		oldSize := order.Size
		decrSize := oldSize.Sub(newSize)
		order.Size = newSize

		var removed bool
		if order.Size.IsZero() {
			delete(s.Orders, order.OrderId)
			removed = true
		}

		val, found := priceLevels.Get(price)
		if !found {
			panic(fmt.Sprintf("%v %v %v %v", orderId, price, newSize, side))
		}

		changedLevel = val.(*PriceLevel)
		changedLevel.Size = changedLevel.Size.Sub(decrSize)
		if changedLevel.Size.IsZero() {
			priceLevels.Remove(price)
		} else if removed {
			changedLevel.OrderCount--
		}
	}

	s.LogOffset = logOffset
	s.LogSeq = logSeq
	s.Seq++
	return &Level2Change{
		ProductId: s.ProductId,
		Seq:       s.Seq,
		Side:      side.String(),
		Price:     changedLevel.Price.String(),
		Size:      changedLevel.Size.String(),
	}
}

func (s *OrderBook) SnapshotLevel2(levels int) *OrderBookLevel2Snapshot {
	snapshot := OrderBookLevel2Snapshot{
		ProductId: s.ProductId,
		Seq:       s.Seq,
		Asks:      make([][3]interface{}, utils.MinInt(levels, s.Depths[entities.SideSell].Size())),
		Bids:      make([][3]interface{}, utils.MinInt(levels, s.Depths[entities.SideBuy].Size())),
	}

	for itr, i := s.Depths[entities.SideBuy].Iterator(), 0; itr.Next() && i < levels; i++ {
		v := itr.Value().(*PriceLevel)
		snapshot.Bids[i] = [3]interface{}{v.Price.String(), v.Size.String(), v.OrderCount}
	}

	for itr, i := s.Depths[entities.SideSell].Iterator(), 0; itr.Next() && i < levels; i++ {
		v := itr.Value().(*PriceLevel)
		snapshot.Asks[i] = [3]interface{}{v.Price.String(), v.Size.String(), v.OrderCount}
	}

	return &snapshot
}

func (s *OrderBook) SnapshotFull() *OrderBookFullSnapshot {
	snapshot := OrderBookFullSnapshot{
		ProductId: s.ProductId,
		Seq:       s.Seq,
		LogOffset: s.LogOffset,
		LogSeq:    s.LogSeq,
		Orders:    make([]matching.BookOrder, len(s.Orders)),
	}

	i := 0
	for _, order := range s.Orders {
		snapshot.Orders[i] = *order
		i++
	}
	return &snapshot
}

func (s *OrderBook) Restore(snapshot *OrderBookFullSnapshot) {
	for _, order := range snapshot.Orders {
		s.SaveOrder(0, 0, order.OrderId, order.Size, order.Price, order.Side)
	}
	s.ProductId = snapshot.ProductId
	s.Seq = snapshot.Seq
	s.LogOffset = snapshot.LogOffset
	s.LogSeq = snapshot.LogSeq
}

// Redis snapshotstore used to manage snapshots
type RedisSnapshotStore struct {
	RedisClient *redis.Client
}

var store *RedisSnapshotStore
var onceStore sync.Once

func sharedSnapshotStore() *RedisSnapshotStore {
	onceStore.Do(func() {
		gexConfig := conf.GetConfig()

		redisClient := redis.NewClient(&redis.Options{
			Addr:     gexConfig.Redis.Addr,
			Password: gexConfig.Redis.Password,
			DB:       0,
		})

		store = &RedisSnapshotStore{RedisClient: redisClient}
	})

	return store
}

func (s *RedisSnapshotStore) StoreLevel2(productId string, snapshot *OrderBookLevel2Snapshot) error {
	buf, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return s.RedisClient.Set(ORDER_BOOK_L2_SNAPSHOT_KEY_PREFIX+productId, buf, 7*24*time.Hour).Err()
}

func (s *RedisSnapshotStore) GetLastLevel2(productId string) (*OrderBookLevel2Snapshot, error) {
	ret, err := s.RedisClient.Get(ORDER_BOOK_L2_SNAPSHOT_KEY_PREFIX + productId).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var snapshot OrderBookLevel2Snapshot
	err = json.Unmarshal(ret, &snapshot)
	return &snapshot, err
}

func (s *RedisSnapshotStore) StoreFull(productId string, snapshot *OrderBookFullSnapshot) error {
	buf, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return s.RedisClient.Set(ORDER_BOOK_FULL_SNAPSHOT_KEY_PREFIX+productId, buf, 7*24*time.Hour).Err()
}

func (s *RedisSnapshotStore) GetLastFull(productId string) (*OrderBookFullSnapshot, error) {
	ret, err := s.RedisClient.Get(ORDER_BOOK_FULL_SNAPSHOT_KEY_PREFIX + productId).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var snapshot OrderBookFullSnapshot
	err = json.Unmarshal(ret, &snapshot)
	return &snapshot, err
}
