package cmd

import (
	
	"net/http"
	_ "net/http/pprof"
	"strconv"

	//"fmt"

	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/events"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/publisher"
	"github.com/irononet/go-exchange/restapi"
	"github.com/irononet/go-exchange/service"
	"github.com/irononet/go-exchange/worker"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gex",
	Short: "Gex is a cryptocurrency exchange software backend built with Go",
	Long: `Gex (Golang-Exchange) is a crypto currency exchange software backend API built
	with Go, that facilitate the exchange, purchase of various crypto currencies. 
	The exchange is designed to be fast and scalable.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}


// TODO: Finish this later
var runServer = &cobra.Command{
	Use: "runserver", 
	Short: "Starts a new instance of the gex server", 
	Long: `Starting a gex web server at port -p`, 
	
}

// TODO: Finish this later 
func init(){
	//rootCmd.AddCommand()
}

func run() error {
	gexConfig := conf.GetConfig()

	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	

	go events.NewBinLogStream().Start()

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
