package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interfaces/msg/config"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils"
	msg "github.com/cossim/coss-server/services/msg/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"go.uber.org/zap"
	"net/http"
	"sync"
)

var (
	wsRid    int64 = 0            //全局客户端id
	wsMutex        = sync.Mutex{} //锁
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	pool = make(map[string][]*client)
)

type client struct {
	Conn *websocket.Conn
	Uid  string //客户端所有者
	Rid  int64  //客户端id
}

// @Summary websocket请求
// @Description websocket请求
// @Router /msg/ws [get]
func ws(c *gin.Context) {
	//升级http请求为websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	//用户上线
	wsRid++
	client := &client{
		Conn: conn,
		Uid:  "",
		Rid:  wsRid,
	}
	token := c.GetHeader("Authorization")
	if token != "" {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			return
		}
		client.Uid = c2.UserId
	}
	//保存到线程池
	client.wsOnlineClients()
	//读取客户端消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			//用户下线
			client.wsOfflineClients()
			return
		}
	}
}

type SendUserMsgRequest struct {
	SenderId   string `json:"sender_id" binding:"required"`
	ReceiverId string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       uint   `json:"type" binding:"required"`
	ReplayId   uint   `json:"replay_id" binding:"required"`
}

// @Summary 发送私聊消息
// @Description 发送私聊消息
// @Accept  json
// @Produce  json
// @param request body SendUserMsgRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /msg/send/user [post]
func sendUserMsg(c *gin.Context) {
	req := new(SendUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}
	_, err := msgClient.SendUserMessage(context.Background(), &msg.SendUserMsgRequest{
		SenderId:   req.SenderId,
		ReceiverId: req.ReceiverId,
		Content:    req.Content,
		Type:       int32(req.Type),
		ReplayId:   uint64(req.ReplayId),
	})
	if err != nil {
		c.Error(err)
		return
	}
	if _, ok := pool[req.ReceiverId]; ok {
		if len(pool[req.ReceiverId]) > 0 {
			sendWsUserMsg(req.SenderId, req.ReceiverId, req.Content, req.Type)
		}
	}
	response.Success(c, "发送成功", gin.H{})
}

type wsMsg struct {
	Uid   string             `json:"uid"`
	Event config.WSEventType `json:"event"`
	Rid   int64              `json:"rid"`
	Data  interface{}        `json:"data"`
}

// 推送私聊消息
func sendWsUserMsg(senderId, receiverId string, msg string, msgType uint) {
	fmt.Println("待发送消息为：", msg)
	type wsUserMsg struct {
		SenderId string `json:"uid"`
		Content  string `json:"content"`
		MsgType  uint   `json:"msgType"`
	}
	//遍历该用户所有客户端
	for _, c := range pool[receiverId] {
		m := wsMsg{Uid: receiverId, Event: config.SendMessageEvent, Rid: c.Rid, Data: &wsUserMsg{senderId, msg, msgType}}
		js, _ := json.Marshal(m)
		err := c.Conn.WriteMessage(websocket.TextMessage, js)
		if err != nil {
			return
		}
	}

}

// 用户上线
func (c client) wsOnlineClients() {
	fmt.Println(c.Uid, "上线了")
	wsMutex.Lock()
	pool[c.Uid] = append(pool[c.Uid], &c)
	wsMutex.Unlock()
	//通知前端接收离线消息
	fmt.Println("通知前端接收离线消息")
	msg := wsMsg{Uid: c.Uid, Event: config.OnlineEvent, Rid: c.Rid}
	js, _ := json.Marshal(msg)
	//上线推送消息
	c.Conn.WriteMessage(websocket.TextMessage, js)
}

// 用户离线
func (c client) wsOfflineClients() {
	//已存在
	wsMutex.Lock()
	//删除map中指定key的元素
	delete(pool, c.Uid)
	wsMutex.Unlock()
	fmt.Println(c.Uid, "下线了")
	return

}
