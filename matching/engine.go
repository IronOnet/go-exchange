package matching

import (
	"strconv"
	"time"

	"github.com/irononet/go-exchange/entities"
	logger "github.com/siddontang/go-log/log"
)

type Engine struct {

	// productID
	productId string

	OrderBook *OrderBook

	OrderReader OrderReader

	OrderOffset int64

	OrderCh chan *OffsetOrder

	LogStore LogStore

	LogCh chan Log

	SnapshotReqCh chan *Snapshot

	SnapshotApproveReqCh chan *Snapshot

	SnapshotCh chan *Snapshot

	SnapShotStore SnapshotStore
}

type Snapshot struct {
	OrderBookSnapshot orderBooSnapShot
	OrderOffset       int64
}

type OffsetOrder struct {
	Offset int64
	Order  *entities.Order
}

func NewEngine(product *entities.Product, orderReader OrderReader, logStore LogStore, snapshotStore SnapshotStore) *Engine {
	e := &Engine{
		productId:            strconv.Itoa(int(product.ID)),
		OrderBook:            NewOrderBook(product),
		LogCh:                make(chan Log, 10000),
		OrderCh:              make(chan *OffsetOrder, 10000),
		SnapshotReqCh:     make(chan *Snapshot, 32),
		SnapshotApproveReqCh: make(chan *Snapshot, 32),
		SnapshotCh:           make(chan *Snapshot, 32),
		SnapShotStore:        snapshotStore,
		OrderReader:          orderReader,
		LogStore:             logStore,
	}

	snapshot, err := snapshotStore.GetLatest()
	if err != nil {
		logger.Fatalf("get latest snaphost error: %v", err)
	}

	if snapshot != nil {
		e.restore(snapshot)
	}
	return e
}

func (e *Engine) Start() {
	go e.runFetcher() 
	go e.runApplier() 
	go e.runCommitter() 
	go e.runShapshots()
}

func (e *Engine) runFetcher() {
	var offset = e.OrderOffset
	if offset > 0 {
		offset = offset + 1
	}
	err := e.OrderReader.SetOffset(offset)
	if err != nil {
		logger.Fatalf("set order reader offset error: %v", err)
	}

	for {
		offset, order, err := e.OrderReader.FetchOrder()
		if err != nil {
			logger.Error(err)
			continue
		}
		e.OrderCh <- &OffsetOrder{offset, order}
	}
}

func (e *Engine) runApplier() {
	var orderOffset int64

	for {
		select {
		case offsetOrder := <-e.OrderCh:
			var logs []Log
			if offsetOrder.Order.Status == entities.OrderStatusCancelling {
				logs = e.OrderBook.CancelOrder(offsetOrder.Order)
			} else {
				logs = e.OrderBook.ApplyOrder(offsetOrder.Order)
			}

			for _, log := range logs {
				e.LogCh <- log
			}

			orderOffset = offsetOrder.Offset

		case snapshot := <-e.SnapshotCh:
			delta := orderOffset - snapshot.OrderOffset
			if delta <= 1000 {
				continue
			}

			logger.Infof("should take snaphshot: %v %v-[%v]-%v->", e.productId, snapshot.OrderOffset, delta, orderOffset)

			snapshot.OrderBookSnapshot = e.OrderBook.Snapshot()
			snapshot.OrderOffset = orderOffset
			e.SnapshotApproveReqCh <- snapshot
		}
	}
}

func (e *Engine) runCommitter() {
	var seq = e.OrderBook.LogSeq
	var pending *Snapshot = nil
	var logs []interface{} 

	for{
		select{
		case log := <- e.LogCh: 
			if log.GetSeq() <= seq{
				logger.Infof("discard log seq=%v", seq) 
				continue 
			}

			seq = log.GetSeq() 
			logs = append(logs, log) 

			// chan is not empty and buffer is not full, continue read 
			if len(e.LogCh) > 0 && len(logs) < 100{
				continue 
			}

			// store log, clean buffer 
			err := e.LogStore.Store(logs) 
			if err != nil{
				panic(err) 
			}
			logs = nil 

			// approve pending snapshot 
			if pending != nil && seq >= pending.OrderBookSnapshot.LogSeq{
				e.SnapshotCh <- pending 
				pending = nil 
			}

		case snapshot := <- e.SnapshotApproveReqCh: 
			if seq >= snapshot.OrderBookSnapshot.LogSeq{
				e.SnapshotCh <- snapshot 
				pending = nil 
				continue 
			}

			if pending != nil{
				logger.Infof("discard snapshot request (seq=%v), new one (seq=%v) received", 
				pending.OrderBookSnapshot.LogSeq, snapshot.OrderBookSnapshot.LogSeq)
			}
			pending = snapshot 
		}
	}
}

func (e *Engine) runShapshots(){

	orderOffset := e.OrderOffset 

	for{
		select{
		case <- time.After(30 * time.Second): 
			// make a new snapshot request 
			e.SnapshotReqCh <- &Snapshot{
				OrderOffset: orderOffset, 
			}
		case snapshot := <- e.SnapshotCh: 
			// store snapshot 
			err := e.SnapShotStore.Store(snapshot) 
			if err != nil{
				logger.Warnf("store snapshot failed: %v", err) 
				continue 
			}
			logger.Infof("new snapshot stored :product=%v OrderOffset=%v LogSeq=%v", 
			e.productId, snapshot.OrderOffset, snapshot.OrderBookSnapshot.LogSeq) 

			// update offset for next snapshot request 
			orderOffset = snapshot.OrderOffset 
		}
	}
}

func (e *Engine) restore(snapshot *Snapshot) {
	logger.Infof("restoring: %+v", *snapshot)
	e.OrderOffset = snapshot.OrderOffset
	e.OrderBook.Restore(&snapshot.OrderBookSnapshot)
}
