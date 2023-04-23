package publisher

import (
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/utils"
	"github.com/shopspring/decimal"
	"time"
)

type MatchStream struct {
	ProductId string
	Sub       *Subscription
	BestBid   decimal.Decimal
	BestAsk   decimal.Decimal
	Tick24h   *entities.Tick
	Tick30d   *entities.Tick
	LogReader matching.LogReader
}

func NewMatchStream(productId string, sub *Subscription, logReader matching.LogReader) *MatchStream {
	s := &MatchStream{
		ProductId: productId,
		Sub:       sub,
		LogReader: logReader,
	}

	s.LogReader.RegisterObserver(s)
	return s
}

func (s *MatchStream) Start() {
	// -1: read from end
	go s.LogReader.Run(0, -1)
}

func (s *MatchStream) OnOpenLOg(log *matching.OpenLog, offset int64) {
	// do nothing
}

func (s *MatchStream) OnDoneLog(log *matching.DoneLog, offset int64) {
	// do nothing
}

func (s *MatchStream) OnMatchLog(log *matching.MatchLog, offset int64) {
	// push match
	s.Sub.Publish(string(CHANNEL_MATCH.FormatWithProductId(string(log.ProductId))), &MatchMessage{
		Type:         "match",
		TradeId:      log.TradeId,
		Sequence:     log.Sequence,
		Time:         log.Time.Format(time.RFC3339),
		ProductId:    string(log.ProductId),
		Price:        log.Price.String(),
		Side:         log.Side.String(),
		MakerOrderId: utils.I64ToA(log.MakerOrderId),
		TakerOrderId: utils.I64ToA(log.TakerOrderId),
		Size:         log.Size.String(),
	})
}
