package matching

import (
	"strconv"

	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/service"
	"github.com/siddontang/go-log/log"
)

func StartEngine() {
	gexConfig := conf.GetConfig()

	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		productId := strconv.Itoa(int(product.ID))
		orderReader := NewKafkaOrderReader(productId, gexConfig.Kafka.Brokers)
		snapshotStore := NewRedisSnapShotStore(productId)
		logStore := NewKafkaLogStore(productId, gexConfig.Kafka.Brokers)
		matchEngine := NewEngine(product, orderReader, logStore, snapshotStore)
		matchEngine.Start()
	}

	log.Info("match engine ok")
}
