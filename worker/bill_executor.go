package worker

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/service"
	"github.com/siddontang/go-log/log"
	"time"
)

const workersNum = 10

type BillExecutor struct {
	WorkerChs [workersNum]chan *entities.Bill
}

func NewBillExecutor() *BillExecutor {
	f := &BillExecutor{
		WorkerChs: [workersNum]chan *entities.Bill{},
	}

	for i := 0; i < workersNum; i++ {
		f.WorkerChs[i] = make(chan *entities.Bill, 256)
		go func(idx int) {
			for {
				select {
				case bill := <-f.WorkerChs[idx]:
					err := service.ExecuteBill(bill.UserId, bill.Currency)
					if err != nil {
						log.Error(err)
					}
				}
			}
		}(i)
	}
	return f
}

func (s *BillExecutor) Start() {
	go s.runMqListener()
	go s.runInspector()
}

func (s *BillExecutor) runMqListener() {
	gexConfig := conf.GetConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     gexConfig.Redis.Addr,
		Password: gexConfig.Redis.Password,
		DB:       0,
	})

	for {
		ret := redisClient.BRPop(time.Second*1000, entities.TopicBill)
		if ret.Err() != nil {
			log.Error(ret.Err())
			continue
		}

		var bill entities.Bill
		err := json.Unmarshal([]byte(ret.Val()[1]), &bill)
		if err != nil {
			panic(ret.Err())
		}

		s.WorkerChs[bill.UserId%workersNum] <- &bill
	}
}

func (s *BillExecutor) runInspector() {
	for {
		select {
		case <-time.After(1 * time.Second):
			bills, err := service.GetUnsettledBills()
			if err != nil {
				log.Error(err)
				continue
			}

			for _, bill := range bills {
				s.WorkerChs[bill.UserId%workersNum] <- bill
			}
		}
	}
}
