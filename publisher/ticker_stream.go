package publisher

import (
	"strconv"
	"sync"
	"time"

	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/matching"
	service "github.com/irononet/go-exchange/service"
	"github.com/shopspring/decimal"
	logger "github.com/siddontang/go-log/log"
)

const intervalSec = 3

var lastTickers sync.Map

type TickerStream struct {
	ProductId      string
	Sub            *Subscription
	BestBid        decimal.Decimal
	BestAsk        decimal.Decimal
	LogReader      matching.LogReader
	LastTickerTime int64
}

func NewTickerStream(productId string, sub *Subscription, logReader matching.LogReader) *TickerStream {
	s := &TickerStream{
		ProductId:      productId,
		Sub:            sub,
		LogReader:      logReader,
		LastTickerTime: time.Now().Unix() - intervalSec,
	}

	s.LogReader.RegisterObserver(s)
	return s
}

func (s *TickerStream) Start() {
	go s.LogReader.Run(0, -1)
}

func (s *TickerStream) OnOpenLOg(log *matching.OpenLog, offset int64) {
	// do nothing
}

func (s *TickerStream) OnDoneLog(log *matching.DoneLog, offset int64) {
	// do nothing
}

func (s *TickerStream) OnMatchLog(log *matching.MatchLog, offset int64) {
	if time.Now().Unix()-s.LastTickerTime > intervalSec {
		ticker, err := s.newTickerMessage(log)
		if err != nil {
			logger.Error(err)
			return
		}

		if ticker == nil {
			return
		}

		lastTickers.Store(log.ProductId, ticker)
		s.Sub.Publish(CHANNEL_TICKER.FormatWithProductId(strconv.Itoa(int(log.ProductId))), ticker)
		s.LastTickerTime = time.Now().Unix()
	}
}

func (s *TickerStream) newTickerMessage(log *matching.MatchLog) (*TickerMessage, error) {
	ticks24h, err := service.GetTicksByProductId(s.ProductId, 1*60, 24)
	if err != nil {
		return nil, err
	}
	tick24h := mergeTicks(ticks24h)
	if tick24h == nil {
		tick24h = &entities.Tick{}
	}

	ticks30d, err := service.GetTicksByProductId(s.ProductId, 24*60, 30) 
	if err != nil{
		return nil, err 
	}
	tick30d := mergeTicks(ticks30d) 
	if tick30d == nil{
		tick30d = &entities.Tick{} 
	}

	return &TickerMessage{
		Type: "ticker", 
		TradeId: log.TradeId, 
		Sequence: log.Sequence, 
		Time: log.Time.Format(time.RFC3339), 
		ProductId: strconv.Itoa(int(log.ProductId)), 
		Price: log.Price.String(), 
		Side: log.Side.String(), 
		LastSize: log.Size.String(), 
		Open24h: tick24h.Open.String(),
		Low24h: tick24h.Low.String(), 
		Volume24h: tick24h.Volume.String(), 
		Volume30d: tick30d.Volume.String(),
	}, nil 
}

func mergeTicks(ticks []*entities.Tick) *entities.Tick{
	var t *entities.Tick 
	for i := range ticks{
		tick := ticks[len(ticks)-1-i] 
		if t == nil{
			t = tick 
		} else{
			t.Close = tick.Close 
			t.Low = decimal.Min(t.Low, tick.Low) 
			t.High = decimal.Max(t.High, tick.High) 
			t.Volume = t.Volume.Add(tick.Volume)
		}
	}
	return t 
}

func getLastTicker(productId string) *TickerMessage{
	ticker, found := lastTickers.Load(productId) 
	if !found{
		return nil 
	}
	return ticker.(*TickerMessage)
}
