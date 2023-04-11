package publisher

import (
	"strconv"

	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/service"
	"github.com/siddontang/go-log/log"
)

func StartServer(){
	gexConfig := conf.GetConfig() 

	sub := NewSubscription() 

	NewRedisStream(sub).Start() 

	products, err := service.GetProducts() 
	if err != nil{
		panic(err)
	}

	for _, product := range products{
		productIdStr := strconv.Itoa(int(product.ID))
		NewTickerStream(productIdStr, sub, matching.NewKafkaLogReader("tickerStream", productIdStr, gexConfig.Kafka.Brokers)).Start() 
		NewMatchStream(productIdStr, sub, matching.NewKafkaLogReader("matchStream", productIdStr, gexConfig.Kafka.Brokers)).Start() 
		NewOrderBookStream(productIdStr, sub, matching.NewKafkaLogReader("orderBookStream", productIdStr, gexConfig.Kafka.Brokers)).Start() 

		go NewServer(gexConfig.PushServer.Addr, gexConfig.PushServer.Path, sub).Run() 

		log.Info("websocket server ok")
	}
}