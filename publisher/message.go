package publisher

import "fmt"

type Level2Type string
type Channel string

func (t Channel) Format(productId string, userId int64) string {
	return fmt.Sprintf("%v:%v:%v", t, productId, userId)
}

func (t Channel) FormatWithUserId(userId int64) string {
	return fmt.Sprintf("%v:%v", t, userId)
}

func (t Channel) FormatWithProductId(productId string) string {
	return fmt.Sprintf("%v:%v", t, productId)
}

const (
	LEVEL_2_TYPE_SNAPSHOT = Level2Type("snapshot")
	LEVEL_2_TYPE_UPDATE   = Level2Type("l2update")

	CHANNEL_TICKER  = Channel("ticker")
	CHANNEL_MATCH   = Channel("match")
	CHANNEL_LEVEL_2 = Channel("level2")
	CHANNEL_FUNDS   = Channel("funds")
	CHANNEL_ORDER   = Channel("order")
)

type Request struct {
	Type        string   `json:"type"`
	ProductIds  []string `json:"product_ids"`
	CurrencyIds []string `json:"currency_ids"`
	Channels    []string `json"channels"`
	Token       string   `json:"token"`
}

type Response struct {
	Type       string   `json:"type"`
	ProductIds []string `json:"product_ids"`
	Channels   []string `json:"channels"`
	Token      string   `json:"token"`
}

type Level2SnapshotMessage struct {
	Type      Level2Type       `json:"type"`
	ProductId string           `json:"productId"`
	Bids      [][3]interface{} `json:"bids"`
	Asks      [][3]interface{} `json:"asks"`
}

type Level2UpdateMessage struct {
	Type      Level2Type       `json:"type"`
	ProductId string           `json:"procduct_id"`
	Changes   [][3]interface{} `json:"changes"` // ["buy", "6500.09", "0.847023376"]
}

type Level2Change struct {
	Seq       int64
	ProductId string
	Side      string
	Price     string
	Size      string
}

type MatchMessage struct {
	Type         string `json:"type"`
	TradeId      int64  `json:"tradeId"`
	Sequence     int64  `json:"sequence"`
	Time         string `json:"time"`
	ProductId    string `json:"productId"`
	Price        string `json:"price"`
	Size         string `json:"size"`
	MakerOrderId string `json:"makerOrderId"`
	TakerOrderId string `json:"takerOrderId"`
	Side         string `json:"side"`
}

type TickerMessage struct {
	Type      string `json:"type"`
	TradeId   int64  `json:"tradeId"`
	Sequence  int64  `json:"sequence"`
	Time      string `json:"time"`
	ProductId string `json:"productId"`
	Price     string `json:"price"`
	Side      string `json:"side"`
	LastSize  string `json:"lastSize"`
	BestBid   string `json:"bestBid"`
	BestAsk   string `json:"bestAsk"`
	Volume24h string `json:"volume24h"`
	Volume30d string `json:"volume30d"`
	Low24h    string `json:"low24h"`
	Open24h   string `json:"open24h"`
}

type FundsMessage struct {
	Type      string `json:"type"`
	Sequence  int64  `json:"sequence"`
	UserId    string `json:"userId"`
	Currency  string `json:"currencyCode"`
	Available string `json:"available"`
	Hold      string `json:"hold"`
}

type OrderMessage struct {
	UserId        int64  `json:"userId"`
	Type          string `json:"type"`
	Sequence      int64  `json:"sequence"`
	Id            string `json:"id"`
	Price         string `json:"price"`
	Size          string `json:"size"`
	Funds         string `json:"funds"`
	ProductId     string `json:"productId"`
	Side          string `json:"side"`
	OrderType     string `json:"orderType"`
	CreatedAt     string `json:"createdAt"`
	FillFees      string `json:"fillFees"`
	FilledSize    string `json:"filledSize"`
	ExecutedValue string `json:"executedValue"`
	Status        string `json:"status"`
	Settled       bool   `json:"settled"`
}
