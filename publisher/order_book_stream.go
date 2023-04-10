package publisher 

import (
	"fmt" 
	"github.com/irononet/go-exchange/matching" 
	logger "github.com/siddontang/go-log/log" 
	"sync" 
	"time"
)

type OrderBookStream struct{
	ProductId string 
	LogReader matching.LogReader 
	LogCh chan *LogOffset 
	OrderBook *OrderBook 
	Sub *Subscription 
	SnapshotCh chan interface{}
}

type LogOffset struct{
	Log interface{} 
	Offset int64 
}

func NewOrderBookStream(productId string, sub *Subscription, logReader matching.LogReader) *OrderBookStream{
	s := &OrderBookStream{
		ProductId: productId, 
		OrderBook: NewOrderBook(productId), 
		LogCh: make(chan *LogOffset, 1000), 
		Sub: sub, 
		LogReader: logReader, 
		SnapshotCh: make(chan interface{}, 100),
	}

	// try to restore snapshot 
	snapshot, err := sharedSnapshotStore().GetLastFull(productId) 
	if err != nil{
		logger.Fatalf("get snaphost error: %v", err) 
	}
	if snapshot != nil{
		s.OrderBook.Restore(snapshot)
		logger.Infof("%v order book snaphot loaded: %+v", s.ProductId, snapshot)
	}

	s.LogReader.RegisterObserver(s) 
	return s
}

func (s *OrderBookStream) Start(){
	logOffset := s.OrderBook.LogOffset 
	if logOffset > 0{
		logOffset++
	}

	go s.LogReader.Run(s.OrderBook.LogSeq, logOffset) 
	go s.runApplier() 
	go s.runSnapshots()
}

func (s *OrderBookStream) OnOpenLOg(log *matching.OpenLog, offset int64){
	s.LogCh <- &LogOffset{log, offset} 
}

func (s *OrderBookStream) OnMatchLog(log *matching.MatchLog, offset int64){
	s.LogCh <- &LogOffset{log, offset}
}

func (s *OrderBookStream) OnDoneLog(log *matching.DoneLog, offset int64){
	s.LogCh <- &LogOffset{log, offset}
}

var lastLevel2Snapshots *sync.Map

func (s *OrderBookStream) runApplier(){
	var lastLevel2Snapshot *OrderBookLevel2Snapshot 
	var lastFullSnapshot *OrderBookFullSnapshot 

	for{
		select{
		case logOffset := <- s.LogCh: 
			var l2Change *Level2Change 

			switch logOffset.Log.(type){
			case *matching.DoneLog: 
				log := logOffset.Log.(*matching.DoneLog) 
				order, found := s.OrderBook.Orders[log.OrderId] 
				if !found{
					continue 
				}
				newSize := order.Size.Sub(log.RemainingSize) 
				l2Change = s.OrderBook.SaveOrder(logOffset.Offset, log.Sequence, log.OrderId, newSize, log.Price, log.Side) 

			case *matching.OpenLog: 
				log := logOffset.Log.(*matching.OpenLog) 
				l2Change = s.OrderBook.SaveOrder(logOffset.Offset, log.Sequence, log.OrderId, log.RemainingSize, log.Price, log.Side) 

			case *matching.MatchLog: 
				log := logOffset.Log.(*matching.MatchLog) 
				order, found := s.OrderBook.Orders[log.MakerOrderId] 
				if !found{
					panic(fmt.Sprintf("should not happen : %+v", log))
				}
				newSize := order.Size.Sub(log.Size) 
				l2Change = s.OrderBook.SaveOrder(logOffset.Offset, log.Sequence, log.MakerOrderId, newSize, log.Price, log.Side)
			}

			if lastLevel2Snapshot == nil || s.OrderBook.Seq-lastLevel2Snapshot.Seq > 10{
				lastLevel2Snapshot = s.OrderBook.SnapshotLevel2(1000) 
				lastLevel2Snapshots.Store(s.ProductId, lastLevel2Snapshot)
			}

			if lastLevel2Snapshot == nil || s.OrderBook.Seq-lastLevel2Snapshot.Seq > 10{
				lastLevel2Snapshot = s.OrderBook.SnapshotLevel2(1000) 
				lastLevel2Snapshots.Store(s.ProductId, lastLevel2Snapshot)
			}

			if lastFullSnapshot == nil || s.OrderBook.Seq-lastFullSnapshot.Seq > 10000{
				lastFullSnapshot = s.OrderBook.SnapshotFull() 
				s.SnapshotCh <- lastFullSnapshot
			}

			if l2Change != nil{
				s.Sub.Publish(string(CHANNEL_LEVEL_2.FormatWithProductId(s.ProductId)), l2Change)
			}

		case <- time.After(200 *time.Millisecond): 
			if lastLevel2Snapshot == nil || s.OrderBook.Seq > lastLevel2Snapshot.Seq{
				lastLevel2Snapshot = s.OrderBook.SnapshotLevel2(1000) 
				lastLevel2Snapshots.Store(s.ProductId, lastLevel2Snapshot)
			}
		}
	}
}

func (s *OrderBookStream) runSnapshots(){
	for{
		select{
		case snapshot := <- s.SnapshotCh: 
			switch snapshot.(type){
			case *OrderBookLevel2Snapshot: 
				err := sharedSnapshotStore().StoreLevel2(s.ProductId, snapshot.(*OrderBookLevel2Snapshot)) 
				if err != nil{
					logger.Error(err)
				}
			case *OrderBookFullSnapshot: 
				err := sharedSnapshotStore().StoreFull(s.ProductId, snapshot.(*OrderBookFullSnapshot)) 
				if err != nil{
					logger.Error(err)
				}
			}
		}
	}
}

func getLastLevel2Snnapshot(productId string) *OrderBookLevel2Snapshot{
	snapshot, found := lastLevel2Snapshots.Load(productId) 
	if !found{
		return nil 
	}
	return snapshot.(*OrderBookLevel2Snapshot)
}