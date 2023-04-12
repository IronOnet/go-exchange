package worker 

import (
	"github.com/irononet/go-exchange/matching" 
	"github.com/irononet/go-exchange/entities" 
	"github.com/irononet/go-exchange/store/mysql" 
	"github.com/irononet/go-exchange/service" 
	"github.com/siddontang/go-log/log"
	"time"
)

type FillMaker struct{
	FillCh chan *entities.Fill 
	LogReader matching.LogReader 
	LogOffset int64 
	LogSeq int64 
}

func NewFillMaker(logReader matching.LogReader) *FillMaker{
	t := &FillMaker{
		FillCh: make(chan *entities.Fill, 1000), 
		LogReader: logReader,
	}

	lastFill, err := mysql.SharedStore().GetLastFillByProductId(logReader.GetProductId()) 
	if err != nil{
		panic(err)
	}

	if lastFill != nil{
		t.LogOffset = lastFill.LogOffset 
		t.LogSeq = lastFill.LogSeq
	}

	t.LogReader.RegisterObserver(t)
	return t 

}

func (t *FillMaker) Start(){
	if t.LogOffset > 0{
		t.LogOffset++
	}
	go t.LogReader.Run(t.LogSeq, t.LogOffset) 
	go t.flusher()
}

func (t *FillMaker) OnMatchLog(log *matching.MatchLog, offset int64){
	t.FillCh <- &entities.Fill{
		TradeId: log.TradeId, 
		MessageSeq: log.Sequence, 
		OrderId: log.TakerOrderId, 
		ProductId: log.ProductId, 
		Size: log.Size, 
		Price: log.Price, 
		Liquidity: "T", 
		Side: log.Side, 
		LogOffset: offset, 
		LogSeq: log.Sequence, 
	}

	t.FillCh <- &entities.Fill{
		TradeId: log.TradeId, 
		MessageSeq: log.Sequence, 
		OrderId: log.MakerOrderId, 
		ProductId: log.ProductId, 
		Size: log.Size, 
		Price: log.Price, 
		Liquidity: "M", 
		Side: log.Side.Opposite(), 
		LogOffset: offset, 
		LogSeq: log.Sequence,
	}
}

func (t *FillMaker) OnOpenLOg(log *matching.OpenLog, offset int64){
	_, _ = service.UpdateOrderStatus(log.OrderId, entities.OrderStatusNew, entities.OrderStatusOpen)
}

func (t *FillMaker) OnDoneLog(log *matching.DoneLog, offset int64){
	t.FillCh <- &entities.Fill{
		MessageSeq: log.Sequence, 
		OrderId: log.OrderId, 
		ProductId: log.ProductId, 
		Size: log.RemainingSize, 
		Done: true, 
		DoneReason: log.Reason, 
		LogOffset: offset, 
		LogSeq: log.Sequence,
	}
}

func (t *FillMaker) flusher(){
	var fills []*entities.Fill 

	for{
		select{
		case fill := <- t.FillCh: 
			fills = append(fills, fill) 

			if len(t.FillCh) > 0 && len(fills) < 1000{
				continue 
			}

			for{
				err := service.AddFills(fills) 
				if err != nil{
					log.Error(err) 
					time.Sleep(time.Second) 
					continue 
				}
				fills = nil 
				break 
			}
		}
	}
}