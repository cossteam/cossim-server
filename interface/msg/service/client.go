package service

import (
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/pkg/msg_queue"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"sync"
)

type client struct {
	Conn       *websocket.Conn
	Uid        string //客户端所有者
	Rid        int64  //客户端id
	ClientType string //客户端类型
	queue      *amqp091.Channel
	wsMutex    sync.Mutex
}

// 用户上线
func (c *client) wsOnlineClients() {
	wsMutex.Lock()
	pool[c.Uid][c.ClientType] = append(pool[c.Uid][c.ClientType], c)
	wsMutex.Unlock()

	//通知前端接收离线消息
	msg := config.WsMsg{Uid: c.Uid, Event: config.OnlineEvent, Rid: c.Rid, SendAt: pkgtime.Now()}
	js, _ := json.Marshal(msg)
	if Enc == nil {
		log.Println("加密客户端错误", zap.Error(nil))
		return
	}
	message, err := Enc.GetSecretMessage(string(js), c.Uid)
	if err != nil {
		fmt.Println("加密失败：", err)
		return
	}
	//上线推送消息
	c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	for {
		msg, ok, err := msg_queue.ConsumeMessages(c.Uid, c.queue)
		if err != nil || !ok {
			//c.queue.Stop()
			//拉取完之后删除队列
			_ = rabbitMQClient.DeleteEmptyQueue(c.Uid)
			return
		}
		c.Conn.WriteMessage(websocket.TextMessage, msg.Body)
	}
}

func (c *client) wsOfflineClients() {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	if _, ok := pool[c.Uid][c.ClientType]; ok {
		for i, c2 := range pool[c.Uid][c.ClientType] {
			if c2.Rid == c.Rid {
				pool[c.Uid][c.ClientType] = append(pool[c.Uid][c.ClientType][:i], pool[c.Uid][c.ClientType][i+1:]...)
				break
			}
		}
	}
}
