package connect

import (
	"github.com/cossim/coss-server/pkg/long_conn"
	"github.com/cossim/coss-server/pkg/long_conn/ws"
	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	Rid  int64
	Conn long_conn.LongConn
}

// NewWebsocketClient creates a new WebsocketClient instance
func NewWebsocketClient(rid int64, conn *websocket.Conn) *WebsocketClient {
	return &WebsocketClient{
		Rid:  rid,
		Conn: ws.New(0, conn),
	}
}
