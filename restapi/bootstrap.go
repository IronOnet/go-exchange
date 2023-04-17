package restapi 

import (
	"github.com/irononet/go-exchange/conf" 
	"github.com/siddontang/go-log/log" 
)

func StartServer(){
	gexConfig := conf.GetConfig() 

	httpServer := NewHttpServer(gexConfig.RestServer.Addr) 
	go httpServer.Start() 

	log.Info("rest server ok")
}