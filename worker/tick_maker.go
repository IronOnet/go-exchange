package worker 

import (
	"github.com/irononet/go-exchange/matching" 
	"github.com/irononet/go-exchange/entities" 
	"github.com/irononet/go-exchange/service" 
	"github.com/shopspring/decimal" 
	"github.com/siddontang/go-log/log" 
	"time" 
)

var minutes = []int64{1, 3, 5, 15, 30, 60, 120, 240, 360, 720, 1440 } 

type TickMaker struct{
	Ticks map[int64]*entities.Tick 
	TickCh chan entities.Tick 
	LogReader matching.LogReader 
	LogOffset int64 
	LogSeq int64 
}

func NewTickMaker(productId string, logReader matching.LogReader) *TickMaker{
	t := &TickMaker{
		Ticks: map[int64]*entities.Tick{}, 
		TickCh: make(chan entities.Tick, 1000) , 
		LogReader: logReader,
	}

	for _, granularity := range minutes{
		tick, err := service.GetLastTickByProductId(productId, granularity) 
		if err != nil{
			panic(err)
		}
		if tick != nil{
			log.Infof("load last tick: %v", tick) 
			t.Ticks[granularity] = tick 
			t.LogOffset = tick.LogOffset 
			t.LogSeq = tick.LogSeq
		}
	}

	t.LogReader.RegisterObserver(t)
	return t
}

func (t *TickMaker) Start(){
	if t.LogOffset > 0{
		t.LogOffset++
	}

	go t.LogReader.Run(t.LogSeq, t.LogOffset) 
	go t.flusher() 
}

func (t *TickMaker) OnOpenLOg(log *matching.OpenLog, offset int64){
	// do nothing 
}

func (t *TickMaker) OnDoneLog(log *matching.DoneLog, offset int64){
	// do nothing 
}

func (t *TickMaker) OnMatchLog(log *matching.MatchLog, offset int64){
	for _, granularity := range minutes{
		tickTime := log.Time.UTC().Truncate(time.Duration(granularity) * time.Minute).Unix() 

		tick, found := t.Ticks[granularity] 
		if !found || tick.Time != tickTime{
			tick = &entities.Tick{
				Open: log.Price, 
				Close: log.Price, 
				Low: log.Price, 
				High: log.Price, 
				Volume: log.Size, 
				ProductId: log.ProductId, 
				Granularity: granularity, 
				Time: tickTime, 
				LogOffset: offset, 
				LogSeq: log.Sequence, 
			}

			t.Ticks[granularity] = tick 
		} else{
			tick.Close = log.Price 
			tick.Low = decimal.Min(tick.Low, log.Price) 
			tick.High = decimal.Max(tick.High, log.Price) 
			tick.Volume = tick.Volume.Add(log.Size) 
			tick.LogOffset = offset 
			tick.LogSeq = log.Sequence 
		}

		t.TickCh <- *tick 
	}
}

func (t *TickMaker) flusher(){
	var ticks []*entities.Tick 

	for{
		select{
		case tick := <- t.TickCh: 
			ticks = append(ticks, &tick)

			if len(t.TickCh) > 0 && len(ticks) < 1000{
				continue 
			}

			for{
				err := service.AddTicks(ticks) 
				if err != nil{
					log.Error(err) 
					time.Sleep(time.Second) 
					continue 
				}
				ticks = nil 
				break 
			}
		}
	}
}