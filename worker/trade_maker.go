package worker

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/service"
	"github.com/irononet/go-exchange/store/mysql"
	"github.com/prometheus/common/log"
	"time"
)

type TradeMaker struct {
	TradeCh   chan *entities.Trade
	LogReader matching.LogReader
	LogOffset int64
	LogSeq    int64
}

func NewTradeMaker(logReader matching.LogReader) *TradeMaker {
	t := &TradeMaker{
		TradeCh:   make(chan *entities.Trade, 1000),
		LogReader: logReader,
	}

	lastTrade, err := mysql.SharedStore().GetLastTradeByProduct(logReader.GetProductId())
	if err != nil {
		panic(err)
	}
	if lastTrade != nil {
		t.LogOffset = lastTrade.LogOffset
		t.LogSeq = lastTrade.LogSeq
	}

	t.LogReader.RegisterObserver(t)
	return t
}

func (t *TradeMaker) Start() {
	if t.LogOffset > 0 {
		t.LogOffset++
	}

	go t.LogReader.Run(t.LogSeq, t.LogOffset)
	go t.runFlusher()
}

func (t *TradeMaker) OnOpenLOg(log *matching.OpenLog, offset int64) {
	// do nothing
}

func (t *TradeMaker) OnDoneLog(log *matching.DoneLog, offset int64) {
	// do nothing
}

func (t *TradeMaker) OnMatchLog(log *matching.MatchLog, offset int64) {
	t.TradeCh <- &entities.Trade{
		TradeId:      log.TradeId,
		ProductId:    log.ProductId,
		TakerOrderId: log.TakerOrderId,
		MakerOrderId: log.MakerOrderId,
		Price:        log.Price,
		Size:         log.Size,
		Side:         log.Side,
		Time:         log.Time,
		LogOffset:    offset,
		LogSeq:       log.Sequence,
	}
}

func (t *TradeMaker) runFlusher() {
	var trades []*entities.Trade

	for {
		select {
		case trade := <-t.TradeCh:
			trades = append(trades, trade)

			if len(t.TradeCh) > 0 && len(trades) < 1000 {
				continue
			}

			for {
				err := service.AddTrades(trades)
				if err != nil {
					log.Error(err)
					time.Sleep(time.Second)
					continue
				}
				trades = nil
				break
			}
		}
	}
}
