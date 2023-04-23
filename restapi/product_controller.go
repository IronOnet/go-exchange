package restapi

import (
	"github.com/gin-gonic/gin"
	"github.com/irononet/go-exchange/service"
	"github.com/irononet/go-exchange/utils"
	"net/http"
)

// GET /products
func GetProducts(ctx *gin.Context) {
	products, err := service.GetProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	var productVos []*ProductVo
	for _, product := range products {
		productVos = append(productVos, newProductVo(product))
	}

	ctx.JSON(http.StatusOK, productVos)
}

// Get products/<product-id>/book?level=[1, 2, 3]
func GetProductOrderBook(ctx *gin.Context) {

}

// GET /products/<product-id>/ticker
func GetProductTicker() {
	// Todo
}

// GET /products/<product-id>/trades
func GetProductTrades(ctx *gin.Context) {
	productId := ctx.Param("productId")

	var tradeVos []*tradeVo
	trades, _ := service.GetTradesByProductId(productId, 50)
	for _, trade := range trades {
		tradeVos = append(tradeVos, newTradeVo(trade))
	}

	ctx.JSON(http.StatusOK, tradeVos)
}

// GET /product/<product-id>/candles
func GetProductCandles(ctx *gin.Context) {
	productId := ctx.Param("productId")
	granularity, _ := utils.AToInt64(ctx.Query("granularity"))
	limit, _ := utils.AToInt64(ctx.DefaultQuery("limit", "1000"))
	if limit <= 0 || limit > 10000 {
		limit = 100
	}

	// [
	//		[time, low, high, open, close, volume],
	//		[14153987678, 0.32, 4.2, 0.35, 4.2, 12.3]
	// ]

	var tickVos [][6]float64
	ticks, err := service.GetTicksByProductId(productId, granularity/60, int(limit))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}
	for _, tick := range ticks {
		tickVos = append(tickVos, [6]float64{float64(tick.Time), utils.DToF64(tick.Low), utils.DToF64(tick.High),
			utils.DToF64(tick.Open), utils.DToF64(tick.Close), utils.DToF64(tick.Volume)})
	}

	ctx.JSON(http.StatusOK, tickVos)
}
