package matching

import (
	"strconv"

	"github.com/irononet/go-exchange/conf"
	"github.com/siddontang/go-log/log"
	"github.com/irononet/go-exchange/services"
)

func StartEngine(){
	gexConfig := conf.GetConfig() 

	products, err := services.GetProducts() 
	if err != nil{
		panic(err)
	}

	for _, product := range products{
		productId := strconv.Itoa(int(product.ID))
		orderReader := NewKafkaOrderReader(productId, gexConfig.Kafka.Brokers) 
		snapshotStore := NewRedisSnapShotStore(productId) 
		logStore := NewKafkaLogStore(productId, gexConfig.Kafka.Brokers) 
		matchEngine := NewEngine(product, orderReader, logStore, snapshotStore) 
		matchEngine.Start()
	}

	log.Info("match engine ok")
}