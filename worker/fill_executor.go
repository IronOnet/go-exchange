package worker 

import (
	"encoding/json" 
	"github.com/irononet/go-exchange/conf" 
	"github.com/irononet/go-exchange/entities" 
	"github.com/irononet/go-exchange/service" 
	"github.com/go-redis/redis" 
	lru "github.com/hashicorp/golang-lru"
	"github.com/siddontang/go-log/log" 
	"time"
)

const FILL_WORKER_NUM = 10 

type FillExecutor struct{
	WorkerChs [FILL_WORKER_NUM]chan *entities.Fill 
}

func NewFillExecutor() *FillExecutor{
	f := &FillExecutor{
		WorkerChs: [FILL_WORKER_NUM]chan *entities.Fill{},
	}

	for i := 0; i < FILL_WORKER_NUM; i++{
		f.WorkerChs[i] = make(chan *entities.Fill, 512)
		go func(idx int){
			settleOrderCache, err := lru.New(1000)
			if err != nil{
				panic(err)
			}

			for{
				select{
				case fill := <- f.WorkerChs[idx]: 
					if settleOrderCache.Contains(fill.OrderId){
						continue 
					}

					order, err := service.GetOrderById(fill.OrderId) 
					if err != nil{
						log.Error(err)
					}
					if order == nil{
						log.Warnf("order not found: %v", fill.OrderId)
						continue 
					}
					if order.Status == entities.OrderStatusCancelled || order.Status == entities.OrderStatusFilled{
						settleOrderCache.Add(order.ID, struct{}{}) 
						continue
					}

					err = service.ExecuteFill(fill.OrderId)
					if err != nil{
						log.Error(err)
					}
				}
			}
		}(i)
	}
	return f
}

func (s *FillExecutor) Start(){
	go s.runInspector() 
	go s.runMqListener()
}

func (s *FillExecutor) runMqListener(){
	config := conf.GetConfig() 

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Redis.Addr, 
		Password: config.Redis.Password, 
		DB: 0, 
	})

	for{
		ret := redisClient.BRPop(time.Second*1000, entities.TopicFill) 
		if ret.Err() != nil{
			log.Error(ret.Err()) 
			continue 
		}

		var fill entities.Fill 
		err := json.Unmarshal([]byte(ret.Val()[1]), &fill) 
		if err != nil{
			log.Error(err) 
			continue 
		}

		s.WorkerChs[fill.OrderId%FILL_WORKER_NUM] <- &fill 
	}
}

func (s *FillExecutor) runInspector(){
	for{
		select{
		case <- time.After(1 * time.Second): 
			fills, err := service.GetUnsettledFills(1000) 
			if err != nil{
				log.Error(err) 
				continue 
			}

			for _, fill := range fills{
				s.WorkerChs[fill.OrderId%FILL_WORKER_NUM] <- fill 
			}
		}
	}
}