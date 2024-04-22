package connect

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type UserIndex struct {
	UserId    string
	cLock     sync.RWMutex // protect the channels for chs
	WsClients []*WebsocketClient
}

// 构造方法
func NewUserIndex(userID string) *UserIndex {
	return &UserIndex{
		UserId:    userID,
		cLock:     sync.RWMutex{},
		WsClients: make([]*WebsocketClient, 0),
	}
}

func (u *UserIndex) GetLength() int {
	return len(u.WsClients)
}

// push 方法用于将 WebSocket 客户端添加到 UserIndex 中
func (u *UserIndex) Push(client *WebsocketClient) {
	u.cLock.Lock()
	defer u.cLock.Unlock()

	u.WsClients = append(u.WsClients, client)
}

// get 方法用于获取 UserIndex 中的 WebSocket 客户端列表
func (u *UserIndex) Get() []*WebsocketClient {
	u.cLock.RLock()
	defer u.cLock.RUnlock()

	return u.WsClients
}

// DeleteByRid 方法用于根据 Rid 从 UserIndex 中删除对应的 WebSocket 客户端
func (u *UserIndex) DeleteByRid(rid int64) {
	u.cLock.Lock()
	defer u.cLock.Unlock()
	for i, client := range u.WsClients {
		if client.Rid == rid {
			//关闭该链接
			err := client.Conn.Close()
			if err != nil {
				fmt.Println("关闭ws客户端链接失败", err)
				return
			}
			// 通过将切片中对应元素与最后一个元素交换位置，然后缩减切片长度的方式删除元素
			u.WsClients[i] = u.WsClients[len(u.WsClients)-1]
			u.WsClients = u.WsClients[:len(u.WsClients)-1]
			return
		}
	}
}

func (u *UserIndex) SendMessage(message string) error {
	u.cLock.Lock()
	defer u.cLock.Unlock()
	if len(u.WsClients) == 0 {
		return nil
	}
	for _, client := range u.WsClients {
		if client.Conn.IsNil() {
			u.DeleteByRid(client.Rid)
			continue
		}
		err := client.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			// Check if the error is due to connection close
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				u.DeleteByRid(client.Rid)
				continue
			}
			return err

		}
	}
	return nil
}

func (u *UserIndex) DelUserWsClient(rid int64) {
	u.DeleteByRid(rid)
}
