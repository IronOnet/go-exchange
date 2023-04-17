package restapi

import (
	"io"

	"github.com/gin-gonic/gin"
	//"github.com/irononet/go-exchange/conf"
)

type HttpServer struct{
	Addr string 
}

func NewHttpServer(addr string) *HttpServer{
	return &HttpServer{
		Addr: addr,
	}
}

func (server *HttpServer) Start(){
	gin.SetMode(gin.DebugMode) 
	gin.DefaultWriter = io.Discard

	r := gin.Default() 
	r.Use(setCORSOptions)

	r.GET("/api/configs", GetConfigs) 
	r.POST("/api/users", SignUp)
	r.POST("/api/users/accessToken", SignIn) 
	r.POST("/api/users/token", GetToken) 
	r.GET("/api/products", GetProducts) 
	r.GET("/api/products/:productId/trades", GetProductTrades) 
	r.GET("/api/products/:productId/book", GetProductOrderBook) 
	r.GET("/api/products/:productId/candles", GetProductCandles) 

	private := r.Group("/", CheckToken())
	{
		private.GET("/api/orders", GetOrders) 
		private.POST("/api/orders", PlaceOrder) 
		private.DELETE("/api/orders/:orderId", CancelOrder) 
		private.DELETE("/api/orders", CancelOrders) 
		private.GET("/api/accounts", GetAccounts) 
		private.GET("/api/users/self", GetUserSelf) 
		private.POST("/api/users/password", ChangePassword) 
		private.DELETE("/api/users/accessToken", SignOut)
		private.GET("/api/wallets/:currency/address", GetWalletAddress) 
		private.GET("/api/wallets/:currency/transactions", GetWalletTransactions) 
		private.POST("/api/wallets/:currency/withdrawal", Withdrawal)
	}

	err := r.Run(server.Addr) 
	if err != nil{
		panic(err)
	}
}

func setCORSOptions(c *gin.Context){
	//c.Header("Access-control-Allow-Origin", conf.GetConfig().CORS.AllowedOrigins[0]) 
	c.Header("Access-Control-Allow-Origin", "*") 
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS") 
	c.Header("Access-Control-Allo-Headers", "*") 
	c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS") 
	c.Header("Content-Type", "application/json")
}