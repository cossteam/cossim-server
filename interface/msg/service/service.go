package service

import (
	"fmt"
	config2 "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/msg/server/http"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/goccy/go-json"
)

// Service struct
type Service struct {
	mqClient *msg_queue.RabbitMQ
}

func New(c *config.AppConfig) (s *Service) {
	mqClient, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Address))
	if err != nil {
		panic(err)
	}
	return &Service{
		mqClient: mqClient,
	}
}

// 监听服务消息队列
func (s Service) Start() {
	s.ListenQueue()
}

func (s Service) ListenQueue() {
	if s.mqClient.GetConnection().IsClosed() {
		panic("mqClient Connection is closed")
	}
	msgs, err := s.mqClient.ConsumeServiceMessages(msg_queue.MsgService, msg_queue.Service_Exchange)
	if err != nil {
		panic(err)
	}
	go func() {
		//监听队列
		for msg := range msgs {
			var msg_query msg_queue.ServiceQueueMsg
			err := json.Unmarshal(msg.Body, &msg_query)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				return
			}
			switch msg_query.Action {
			case msg_queue.SendMessage:
				mmap, ok := msg_query.Data.(map[string]interface{})
				if !ok {
					fmt.Println("解析消息失败：")
					return
				}
				//map解析成结构体
				jsonData, err := json.Marshal(mmap)
				if err != nil {
					fmt.Println("Failed to marshal map to JSON:", err)
					return
				}
				var wsm config2.WsMsg
				err = json.Unmarshal(jsonData, &wsm)
				if err != nil {
					fmt.Println("解析消息失败：", err)
					return
				}
				http.SendMsg(wsm.Uid, wsm.Event, wsm.Data)
			}
		}
	}()
}
func (s Service) Stop() {
	//关闭队列
	s.mqClient.Close()
}
