package main

import (
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/irononet/go-exchange/conf"
	//"github.com/irononet/go-exchange/events"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/publisher"
	"github.com/irononet/go-exchange/restapi"
	"github.com/irononet/go-exchange/service"
	"github.com/irononet/go-exchange/worker"
	"github.com/prometheus/common/log"
)

func main() {

	gexConfig := conf.GetConfig()

	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	restapi.StartServer()

	//go events.NewBinLogStream().Start()

	matching.StartEngine()

	publisher.StartServer()

	worker.NewFillExecutor().Start()

	worker.NewBillExecutor().Start()

	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		productIdStr := strconv.Itoa(int(product.ID))
		worker.NewTickMaker(productIdStr, matching.NewKafkaLogReader("tickMaker", productIdStr, gexConfig.Kafka.Brokers)).Start()
		worker.NewFillMaker(matching.NewKafkaLogReader("fillMaker", productIdStr, gexConfig.Kafka.Brokers)).Start()
		worker.NewTradeMaker(matching.NewKafkaLogReader("tradeMaker", productIdStr, gexConfig.Kafka.Brokers)).Start()

	}

	restapi.StartServer()

	select {}
}
