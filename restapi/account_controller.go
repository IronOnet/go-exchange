package restapi 

import (
	"github.com/gin-gonic/gin" 
	"github.com/irononet/go-exchange/service" 
	"net/http"
)

func GetAccounts(ctx *gin.Context){
	var accountVos []*accountVo 
	currencies := ctx.QueryArray("currency") 
	if len(currencies) != 0{
		for _, currency := range currencies{
			account, err := service.GetAccount(int64(GetCurrentUser(ctx).ID), currency) 
			if err != nil{
				ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
				return 
			}

			if account == nil{
				continue 
			}

			accountVos = append(accountVos, newAccountVo(account))
		}
	} else{
		accounts, err := service.GetAccountsByUserId(int64(GetCurrentUser(ctx).ID)) 
		if err != nil{
			ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
			return 
		}
		for _, account := range accounts{
			accountVos = append(accountVos, newAccountVo(account))
		}
	}
	ctx.JSON(http.StatusOK, accountVos)
}