package restapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/matching"
	"github.com/irononet/go-exchange/service"
	"github.com/irononet/go-exchange/utils"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
)

var productId2Writer sync.Map

func getWriter(productId string) *kafka.Writer {
	writer, found := productId2Writer.Load(productId)
	if found {
		return writer.(*kafka.Writer)
	}

	gexConfig := conf.GetConfig()

	newWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      gexConfig.Kafka.Brokers,
		Topic:        matching.TopicOrderPrefix + productId,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 5 * time.Millisecond,
	})
	productId2Writer.Store(productId, newWriter)
	return newWriter
}

func submitOrder(order *entities.Order) {
	buf, err := json.Marshal(order)
	if err != nil {
		log.Error(err)
		return
	}

	productIdStr := strconv.Itoa(order.ProductId)
	err = getWriter(productIdStr).WriteMessages(context.Background(), kafka.Message{Value: buf})
	if err != nil {
		log.Error(err)
	}

}

// POST /orders
func PlaceOrder(ctx *gin.Context) {
	var req placeOrderRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}

	side := entities.Side(req.Side)
	if len(side) == 0 {
		side = entities.SideBuy
	}

	orderType := entities.OrderType(req.Type)
	if len(orderType) == 0 {
		orderType = entities.LIMIT_ORDER
	}

	if len(req.ClientOid) > 0 {
		_, err = uuid.Parse(req.ClientOid)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, newMessageVo(fmt.Errorf("invalid client_oid: %v", err)))
			return
		}
	}

	size := decimal.NewFromFloat(req.Size)
	price := decimal.NewFromFloat(req.Price)
	funds := decimal.NewFromFloat(req.Funds)

	order, err := service.PlaceOrder(int64(GetCurrentUser(ctx).ID), req.ClientOid, req.ProductId, orderType, side, size, price, funds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	submitOrder(order)

	ctx.JSON(http.StatusOK, order)
}

// DELETE /orders/1
// DELETE /orders/client:1
func CancelOrder(ctx *gin.Context) {
	rawOrderId := ctx.Param("orderId")

	var order *entities.Order
	var err error
	if strings.HasPrefix(rawOrderId, "client:") {
		clientOid := strings.Split(rawOrderId, ":")[1]
		order, err = service.GetOrderByClientUid(int64(GetCurrentUser(ctx).ID), clientOid)
	} else {
		orderId, _ := utils.AToInt64(rawOrderId)
		order, err = service.GetOrderById(orderId)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	if order == nil || order.UserId != int(GetCurrentUser(ctx).ID) {
		ctx.JSON(http.StatusNotFound, newMessageVo(errors.New("order not found")))
		return
	}

	order.Status = entities.OrderStatusCancelling
	submitOrder(order)

	ctx.JSON(http.StatusOK, nil)
}

// DELETE /orders/?productId=BTC-USD&side=[buy, sell]
func CancelOrders(ctx *gin.Context) {
	productId := ctx.Query("productId")

	var side *entities.Side
	var err error
	rawSide := ctx.Query("side")
	if len(rawSide) > 0 {
		side, err = entities.NewSideFromString(rawSide)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
			return
		}

	}

	orders, err := service.GetOrdersByUserId(int64(GetCurrentUser(ctx).ID), []entities.OrderStatus{entities.OrderStatusOpen, entities.OrderStatusNew}, side, productId, 0, 0, 10000)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	for _, order := range orders {
		order.Status = entities.OrderStatusCancelling
		submitOrder(order)
	}

	ctx.JSON(http.StatusOK, nil)
}

// GET /orders
func GetOrders(ctx *gin.Context) {
	productId := ctx.Query("productId")

	var side *entities.Side
	var err error
	rawSide := ctx.GetString("Side")
	if len(rawSide) > 0 {
		side, err = entities.NewSideFromString(rawSide)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, newMessageVo(err))
			return
		}
	}

	var statuses []entities.OrderStatus
	statusValues := ctx.QueryArray("status")
	for _, statusValue := range statusValues {
		status, err := entities.NewOrderStatusFromString(statusValue)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, newMessageVo(err))
			return
		}
		statuses = append(statuses, *status)
	}

	before, _ := strconv.ParseInt(ctx.Query("before"), 10, 64)
	after, _ := strconv.ParseInt(ctx.Query("after"), 10, 64)
	limit, _ := strconv.ParseInt(ctx.Query("limit"), 10, 64)

	orders, err := service.GetOrdersByUserId(int64(GetCurrentUser(ctx).ID), statuses, side, productId, before, after, int(limit))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	orderVos := []*orderVo{}
	for _, order := range orders {
		orderVos = append(orderVos, newOrderVo(order))

	}

	var newBefore, newAfter int64 = 0, 0
	if len(orders) > 0 {
		newBefore = int64(orders[0].ID)
		newAfter = int64(orders[len(orders)-1].ID)
	}

	ctx.Header("gex-before", strconv.FormatInt(newBefore, 10))
	ctx.Header("gex-after", strconv.FormatInt(newAfter, 10))

	ctx.JSON(http.StatusOK, orderVos)
}
