package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/pkg/long_conn"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

var _ long_conn.LongConn = &JSWebSocket{}

type JSWebSocket struct {
	ConnType int
	Conn     *websocket.Conn
}

func (w *JSWebSocket) CheckHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer w.Conn.Close()

	for {
		select {
		case <-ticker.C:
			// 发送 ping 消息
			err := w.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("Failed to send ping: %v\n", err)
				return
			}
			//log.Println("Ping sent successfully")
		}
	}
}

func (w *JSWebSocket) SetReadDeadline(timeout time.Duration) error {
	return nil
}

func (w *JSWebSocket) SetWriteDeadline(timeout time.Duration) error {
	return nil
}

func (w *JSWebSocket) SetReadLimit(limit int64) {
	w.Conn.SetReadLimit(limit)
}

func (w *JSWebSocket) LocalAddr() string {
	return ""
}

func New(connType int, conn *websocket.Conn) *JSWebSocket {
	return &JSWebSocket{ConnType: connType, Conn: conn}
}

func (w *JSWebSocket) Close() error {
	return w.Conn.Close()
}

func (w *JSWebSocket) WriteMessage(messageType int, message []byte) error {
	return w.Conn.WriteMessage(messageType, message)
}

func (w *JSWebSocket) ReadMessage() (int, []byte, error) {
	messageType, b, err := w.Conn.ReadMessage()
	return messageType, b, err
}

func (w *JSWebSocket) dial(ctx context.Context, urlStr string) (*websocket.Conn, *http.Response, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}
	query := u.Query()
	query.Set("isMsgResp", "true")
	u.RawQuery = query.Encode()
	conn, httpResp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	if httpResp == nil {
		httpResp = &http.Response{
			StatusCode: http.StatusSwitchingProtocols,
		}
	}
	_, data, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("read response error %w", err)
	}
	var apiResp struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
		ErrDlt  string `json:"errDlt"`
	}
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal response error %w", err)
	}
	if apiResp.ErrCode == 0 {
		return conn, httpResp, nil
	}
	httpResp.Body = io.NopCloser(bytes.NewReader(data))
	return conn, httpResp, fmt.Errorf("read response error %d %s %s",
		apiResp.ErrCode, apiResp.ErrMsg, apiResp.ErrDlt)
}

func (w *JSWebSocket) Dial(urlStr string, _ http.Header) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, httpResp, err := w.dial(ctx, urlStr)
	if err == nil {
		w.Conn = conn
	}
	return httpResp, err
}

func (w *JSWebSocket) IsNil() bool {
	if w.Conn != nil {
		return false
	}
	return true
}
