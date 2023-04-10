package publisher 

import (
	"encoding/json" 
	"github.com/irononet/go-exchange/conf" 
	"github.com/irononet/go-exchange/entities" 
	"github.com/irononet/go-exchange/utils" 
	"github.com/go-redis/redis" 
	"github.com/siddontang/go-log/log" 
	"sync" 
	"time"
)

type RedisStream struct{
	Sub *Subscription 
	Mutex sync.Mutex 
}

func NewRedisStream(sub *Subscription) *RedisStream{
	return &RedisStream{
		Sub: sub, 
		Mutex: sync.Mutex{},
	}
}


func (r *RedisStream) Start(){
	gexConfig := conf.GetConfig() 

	redisClient := redis.NewClient(&redis.Options{
		Addr: gexConfig.Redis.Addr, 
		Password: gexConfig.Redis.Password, 
		DB: 0, 
	})

	_, err := redisClient.Ping().Result() 
	if err != nil{
		panic(err)
	}

	go func(){
		for{
			ps := redisClient.Subscribe(entities.TopicOrder)
			_, err := ps.Receive() 
			if err != nil{
				log.Error(err) 
				continue 
			}

			for{
				select{
				case msg := <- ps.Channel(): 
					var order entities.Order 
					err := json.Unmarshal([]byte(msg.Payload), &order)
					if err != nil{
						continue 
					}

					r.Sub.Publish(CHANNEL_ORDER.Format(string(order.ProductId), int64(order.UserId)), OrderMessage{
						UserId: int64(order.UserId), 
						Type: "order", 
						Sequence: 0, 
						Id: utils.I64ToA(int64(order.ID)),
						Price:  order.Price.String(), 
						Size: order.Size.String(), 
						Funds: "0", 
						ProductId: string(order.ProductId), 
						Side: order.Side.String(), 
						OrderType: order.Type.String(), 
						CreatedAt: order.CreatedAt.Format(time.RFC3339), 
						FillFees: order.FillFees.String(), 
						FilledSize: order.FilledSize.String(), 
						ExecutedValue: order.ExecutedValue.String(), 
						Status: string(order.Status), 
						Settled: order.Settled,
					})
				}
			}
		}
	}()

	go func(){
		for {
			ps := redisClient.Subscribe(entities.TopicAccount) 
			_, err := ps.Receive() 
			if err != nil{
				log.Error(err) 
				continue 
			}

			for {
				select {
				case msg := <- ps.Channel(): 
					var account entities.Account 
					err := json.Unmarshal([]byte(msg.Payload), &account) 
					if err != nil{
						continue 
					}

					r.Sub.Publish(CHANNEL_FUNDS.FormatWithUserId(account.UserId), FundsMessage{
						Type: "funds", 
						Sequence: 0, 
						UserId: utils.I64ToA(account.UserId), 
						Currency: account.Currency, 
						Hold: account.Hold.String(), 
						Available: account.Available.String(),
					})
				}
			}
		}
	}() 


}