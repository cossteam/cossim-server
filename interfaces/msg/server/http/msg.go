package http

import (
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interfaces/msg/config"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	pool = make(map[string]*client)
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
	//todo 获取请求头的token,并解析出uuid
	//if api.Token != "" {
	//	client.Uid = api.Userinfo.ID
	//}
	//保存到线程池
	client.wsOnlineClients()
	//读取客户端消息
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			//用户下线
			client.wsOfflineClients()
			return
		}
		//输出消息
		fmt.Println("接收到消息", string(data))
	}
}

type wsMsg struct {
	Uid   string             `json:"uid"`
	Event config.WSEventType `json:"event"`
	Rid   int64              `json:"rid"`
	Data  interface{}        `json:"data"`
}

// 用户上线
func (c client) wsOnlineClients() {
	fmt.Println(c.Uid, "上线了")
	wsMutex.Lock()
	pool[c.Uid] = &c
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
