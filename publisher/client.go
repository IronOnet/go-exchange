package publisher

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/irononet/go-exchange/service"
	"github.com/siddontang/go-log/log"
)

const (
	WRITE_WAIT       = 10 * time.Second
	PONG_WAIT        = 60 * time.Second
	PING_PERIOD      = (PONG_WAIT * 9) / 10
	MAX_MESSAGE_SIZE = 512
)

var Id int64

type Client struct {
	Id         int64
	Conn       *websocket.Conn
	WriteCh    chan interface{}
	L2ChangeCh chan *Level2Change
	Sub        *Subscription
	Channels   map[string]struct{}
	Mu         sync.Mutex
}

func NewClient(conn *websocket.Conn, sub *Subscription) *Client {
	return &Client{
		Id:         atomic.AddInt64(&Id, 1),
		Conn:       conn,
		WriteCh:    make(chan interface{}, 256),
		L2ChangeCh: make(chan *Level2Change, 512),
		Sub:        sub,
		Channels:   map[string]struct{}{},
	}
}

func (c *Client) StartServe() {
	go c.RunReader()
	go c.RunWriter()
}

func (c *Client) RunReader() {
	c.Conn.SetReadLimit(MAX_MESSAGE_SIZE)
	err := c.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	if err != nil {
		log.Error(err)
	}
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	})
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			c.Close()
			break
		}

		var req Request

		err = json.Unmarshal(message, &req)
		if err != nil {
			log.Errorf("bad message: %v %v", string(message), err)
			c.Close()
			break
		}

		c.OnMessage(&req)
	}
}

func (c *Client) RunWriter() {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		cancel()
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	go c.RunL2ChangeWriter(ctx)

	for {
		select {
		case message := <-c.WriteCh:
			switch message.(type) {
			case *Level2Change:
				c.L2ChangeCh <- message.(*Level2Change)
				continue
			}

			err := c.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err != nil {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.Close()
				return
			}

			buf, err := json.Marshal(message)
			if err != nil {
				continue
			}
			err = c.Conn.WriteMessage(websocket.TextMessage, buf)
			if err != nil {
				c.Close()
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				c.Close()
				return
			}
		}
	}
}

func (c *Client) RunL2ChangeWriter(ctx context.Context) {
	type state struct {
		resendSnapshot bool
		changes        []*Level2Change
		lastSeq        int64
	}
	states := map[string]*state{}

	stateOf := func(productId string) *state {
		s, found := states[productId]
		if found {
			return s
		}
		s = &state{
			resendSnapshot: true,
			changes:        nil,
			lastSeq:        0,
		}

		states[productId] = s
		return s
	}

	for {
		select {
		case <-ctx.Done():
			return
		case l2Change := <-c.L2ChangeCh:
			state := stateOf(l2Change.ProductId)

			if state.resendSnapshot || l2Change.Seq == 0 {
				snapshot := getLastLevel2Snnapshot(l2Change.ProductId)
				if snapshot == nil {
					log.Warnf("no snapshot for %v", l2Change.ProductId)
					continue
				}

				if state.lastSeq > snapshot.Seq {
					log.Warnf("last snapshot too old: %v changeSeq=%v snapshotSeq=%v", l2Change.ProductId, state.lastSeq, snapshot.Seq)
					continue
				}

				state.lastSeq = snapshot.Seq
				state.resendSnapshot = false

				c.WriteCh <- &Level2SnapshotMessage{
					Type:      LEVEL_2_TYPE_SNAPSHOT,
					ProductId: l2Change.ProductId,
					Bids:      snapshot.Bids,
					Asks:      snapshot.Asks,
				}
				continue
			}

			if l2Change.Seq <= state.lastSeq {
				log.Infof("discard l2changeSeq=%v snapshotSeq=%v", l2Change.Seq, state.lastSeq)
				continue
			}

			if l2Change.Seq != state.lastSeq+1 {
				log.Infof("l2Change lost newSeq=%v lastSeq=%v", l2Change.Seq, state.lastSeq)
				state.resendSnapshot = true
				state.changes = nil
				state.lastSeq = l2Change.Seq
				if len(c.L2ChangeCh) == 0 {
					c.L2ChangeCh <- &Level2Change{ProductId: l2Change.ProductId}
				}
				continue
			}

			state.lastSeq = l2Change.Seq
			state.changes = append(state.changes, l2Change)

			if len(c.L2ChangeCh) > 0 && len(state.changes) < 10 {
				continue
			}

			updateMsg := &Level2UpdateMessage{
				Type:      LEVEL_2_TYPE_UPDATE,
				ProductId: l2Change.ProductId,
			}
			for _, change := range state.changes {
				updateMsg.Changes = append(updateMsg.Changes, [3]interface{}{change.Side, change.Price, change.Size})
			}
			c.WriteCh <- updateMsg
			state.changes = nil
		}
	}
}

func (c *Client) OnMessage(req *Request) {
	switch req.Type {
	case "subscribe":
		c.OnSub(req.CurrencyIds, req.ProductIds, req.Channels, req.Token)
	case "unsubscribe":
		c.OnUnSub(req.CurrencyIds, req.ProductIds, req.Channels, req.Token)
	default:
	}
}

func (c *Client) OnSub(currencyIds []string, productIds []string, channels []string, token string) {
	user, err := service.CheckToken(token)
	if err != nil {
		log.Error(err)
	}

	var userId int64
	if user != nil {
		userId = int64(user.ID)
	}

	for range currencyIds {
		for _, channel := range channels {
			switch Channel(channel) {
			case CHANNEL_FUNDS:
				c.Subscribe(CHANNEL_FUNDS.FormatWithUserId(userId))
			}
		}
	}

	for _, productId := range productIds {
		for _, channel := range channels {
			switch Channel(channel) {
			case CHANNEL_LEVEL_2:
				if c.Subscribe(CHANNEL_LEVEL_2.FormatWithProductId(productId)) {
					if len(c.L2ChangeCh) == 0 {
						c.L2ChangeCh <- &Level2Change{ProductId: productId}
					}
				}

			case CHANNEL_MATCH:
				c.Subscribe(CHANNEL_MATCH.FormatWithProductId(productId))

			case CHANNEL_TICKER:
				if c.Subscribe(CHANNEL_TICKER.FormatWithProductId(productId)) {
					ticker := getLastTicker(productId)
					if ticker != nil {
						c.WriteCh <- ticker
					}
				}

			case CHANNEL_ORDER:
				c.Subscribe(CHANNEL_ORDER.Format(productId, userId))

			default:
				continue
			}
		}
	}
}

func (c *Client) OnUnSub(currencyIds []string, productIds []string, channels []string, token string) {
	user, err := service.CheckToken(token)
	if err != nil {
		log.Error(err)
	}

	var userId int64
	if user != nil {
		userId = int64(user.ID)
	}

	for range currencyIds {
		for _, channel := range channels {
			switch Channel(channel) {
			case CHANNEL_FUNDS:
				c.Unsubscribe(CHANNEL_FUNDS.FormatWithUserId(userId))
			}
		}
	}

	for _, productId := range productIds {
		for _, channel := range channels {
			switch Channel(channel) {
			case CHANNEL_LEVEL_2:
				c.Unsubscribe(CHANNEL_LEVEL_2.FormatWithProductId(productId))
			case CHANNEL_MATCH:
				c.Unsubscribe(CHANNEL_LEVEL_2.FormatWithProductId(productId))
			case CHANNEL_TICKER:
				c.Unsubscribe(CHANNEL_TICKER.FormatWithProductId(productId))
			case CHANNEL_ORDER:
				c.Unsubscribe(CHANNEL_ORDER.Format(productId, userId))

			default:
				continue
			}
		}
	}
}

func (c *Client) Subscribe(channel string) bool {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	_, found := c.Channels[channel]
	if found {
		return false
	}

	if c.Sub.Subscribe(channel, c) {
		c.Channels[channel] = struct{}{}
		return true
	}
	return false
}

func (c *Client) Unsubscribe(channel string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if c.Sub.Unsubscribe(channel, c) {
		delete(c.Channels, channel)
	}
}

func (c *Client) Close() {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	for channel := range c.Channels {
		c.Sub.Unsubscribe(channel, c)
	}
}
