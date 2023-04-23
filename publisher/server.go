package publisher

import (
	"io"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/siddontang/go-log/log"
)

type Server struct {
	Addr string
	Path string
	Sub  *Subscription
}

func NewServer(addr, path string, sub *Subscription) *Server {
	return &Server{
		Addr: addr,
		Path: path,
		Sub:  sub,
	}
}

func (s *Server) Ws(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error(err)
	}

	NewClient(conn, s.Sub).StartServe()
}

func (s *Server) Run() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	r := gin.Default()
	r.GET(s.Path, s.Ws)
	err := r.Run(s.Addr)
	if err != nil {
		panic(err)
	}
}
