package restapi

import (
	"github.com/gin-gonic/gin"
	"github.com/irononet/go-exchange/service"
	"net/http"
)

func GetConfigs(ctx *gin.Context) {
	configs, err := service.GetConfigs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	m := map[string]string{}
	for _, config := range configs {
		m[config.Key] = config.Value
	}

	ctx.JSON(http.StatusOK, m)
}
