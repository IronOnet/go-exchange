package restapi 

import (
	"github.com/gin-gonic/gin" 
	"net/http"
)

// TODO: Comming soon 
// GET /wallets/{currency}/address 
func GetWalletAddress(ctx *gin.Context){
	ctx.JSON(http.StatusOK, walletAddressVo{Address: "0xWIP"})
}

// GET /wallets/{currency}/transaction 
func GetWalletTransactions(ctx *gin.Context){
	currency := ctx.Param("currency") 

	transactionVos := []*transactionVo{} 
	if currency == "BTC"{
		transactionVos = append(transactionVos, newTransactionVo())
	}

	ctx.JSON(http.StatusOK, transactionVos)
}

// POST /wallets/{currency}/withdrawal 
func Withdrawal(ctx *gin.Context){
	ctx.JSON(http.StatusOK, transactionVo{
		Id: "1", 
		Currency: "BTC", 
		Amount: "0.1", 
	})
}