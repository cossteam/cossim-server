package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/config"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils"
	msg "github.com/cossim/coss-server/service/msg/api/v1"
	relation "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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
	var uid string
	token := c.Query("token")

	//升级http请求为websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	if token == "" {
		id, err := pkghttp.ParseTokenReUid(c)
		if err != nil {
			return
		}
		uid = id
	} else {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			return
		}
		uid = c2.UserId
	}
	if uid == "" {
		return
	}
	//用户上线
	wsRid++
	client := &client{
		Conn: conn,
		Uid:  "",
		Rid:  wsRid,
	}

	client.Uid = uid
	//保存到线程池
	client.wsOnlineClients()
	//读取客户端消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			//用户下线
			client.wsOfflineClients()
			return
		}
	}
}

type SendUserMsgRequest struct {
	ReceiverId string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       uint   `json:"type" binding:"required"`
	ReplayId   uint   `json:"replay_id" `
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
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.Fail(c, err.Error(), nil)
		return
	}

	userRelationStatus, err := relationClient.GetUserRelation(context.Background(), &relation.GetUserRelationRequest{
		UserId:   thisId,
		FriendId: req.ReceiverId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	if userRelationStatus.Status != relation.RelationStatus_RELATION_STATUS_ADDED {
		response.Fail(c, "好友关系不正常", nil)
		return
	}

	_, err = msgClient.SendUserMessage(context.Background(), &msg.SendUserMsgRequest{
		SenderId:   thisId,
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
		fmt.Println("pool[req.ReceiverId] => ", pool[req.ReceiverId])
		if len(pool[req.ReceiverId]) > 0 {
			sendWsUserMsg(thisId, req.ReceiverId, req.Content, req.Type, req.ReplayId)
		}
	}
	response.Success(c, "发送成功", gin.H{})
}

type SendGroupMsgRequest struct {
	GroupId  uint32 `json:"group_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     uint32 `json:"type" binding:"required"`
	ReplayId uint32 `json:"replay_id" binding:"required"`
}

// @Summary 发送群聊消息
// @Description 发送群聊消息
// @Accept  json
// @Produce  json
// @param request body SendGroupMsgRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /msg/send/group [post]
func sendGroupMsg(c *gin.Context) {
	req := new(SendGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	//todo 判断是否在群聊里
	//todo 判断是否被禁言
	_, err = msgClient.SendGroupMessage(context.Background(), &msg.SendGroupMsgRequest{
		GroupId:  req.GroupId,
		Content:  req.Content,
		Type:     req.Type,
		ReplayId: req.ReplayId,
	})
	// 发送成功进行消息推送
	if err != nil {
		c.Error(err)
		return
	}
	//查询群聊所有用户id
	uids, err := userGroupClient.GetUserGroupIDs(context.Background(), &relation.GroupID{
		GroupId: uint32(req.GroupId),
	})
	sendWsGroupMsg(uids.UserIds, thisId, req.GroupId, req.Content, req.Type, req.ReplayId)
	response.Success(c, "发送成功", gin.H{})
}

type msgListRequest struct {
	//GroupId    int64  `json:"group_id" binding:"required"`
	UserId   string `json:"user_id" binding:"required"`
	Type     int32  `json:"type" binding:"required"`
	Content  string `json:"content" binding:"required"`
	PageNum  int    `json:"page_num" binding:"required"`
	PageSize int    `json:"page_size" binding:"required"`
}

// @Summary 获取私聊消息
// @Description 获取私聊消息
// @Accept  json
// @Produce  json
// @Param user_id query string true "用户id"
// @Param type query string true "类型"
// @Param content query string false "消息"
// @Param page_num query int false "页码"
// @Param page_size query int false "页大小"
// @Success		200 {object} utils.Response{}
// @Router /msg/list/user [get]
func getUserMsgList(c *gin.Context) {
	var num = c.Query("page_num")
	var size = c.Query("page_size")
	var id = c.Query("user_id")
	var msgType = c.Query("type")
	var content = c.Query("content")

	if num == "" || size == "" || id == "" {
		response.Fail(c, "参数错误", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.Fail(c, err.Error(), nil)
		return
	}
	pageNum, _ := strconv.Atoi(num)
	pageSize, _ := strconv.Atoi(size)
	mt, _ := strconv.Atoi(msgType)
	if pageNum == 0 || pageSize == 0 {
		response.Fail(c, "参数错误", nil)
		return
	}

	var msgListRequest = &msgListRequest{
		UserId:   id,
		Type:     int32(mt),
		Content:  content,
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	msg, err := msgClient.GetUserMessageList(context.Background(), &msg.GetUserMsgListRequest{
		UserId:   thisId,                //当前用户
		FriendId: msgListRequest.UserId, //好友id
		Content:  msgListRequest.Content,
		Type:     msgListRequest.Type,
		PageNum:  int32(msgListRequest.PageNum),
		PageSize: int32(msgListRequest.PageSize),
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, "获取成功", gin.H{"msg_list": msg})
}

type wsMsg struct {
	Uid   string             `json:"uid"`
	Event config.WSEventType `json:"event"`
	Rid   int64              `json:"rid"`
	Data  interface{}        `json:"data"`
}

// 推送私聊消息
func sendWsUserMsg(senderId, receiverId string, msg string, msgType uint, replayId uint) {
	type wsUserMsg struct {
		SenderId string `json:"uid"`
		Content  string `json:"content"`
		MsgType  uint   `json:"msgType"`
		ReplayId uint   `json:"reply_id"`
	}
	//遍历该用户所有客户端
	for _, c := range pool[receiverId] {
		m := wsMsg{Uid: receiverId, Event: config.SendUserMessageEvent, Rid: c.Rid, Data: &wsUserMsg{senderId, msg, msgType, replayId}}
		js, _ := json.Marshal(m)
		err := c.Conn.WriteMessage(websocket.TextMessage, js)
		if err != nil {
			logger.Error("send msg err", zap.Error(err))
			return
		}
	}
}

// 推送群聊消息
func sendWsGroupMsg(uIds []string, userId string, groupId uint32, msg string, msgType uint32, replayId uint32) {
	type wsGroupMsg struct {
		GroupId  int64  `json:"group_id"`
		UserId   string `json:"uid"`
		Content  string `json:"content"`
		MsgType  uint   `json:"msgType"`
		ReplayId uint   `json:"reply_id"`
	}
	//发送群聊消息
	for _, uid := range uIds {
		////遍历该用户所有客户端
		//if uid == userId {
		//	continue
		//}
		for _, c := range pool[uid] {
			m := wsMsg{Uid: uid, Event: config.SendGroupMessageEvent, Rid: c.Rid, Data: &wsGroupMsg{int64(groupId), userId, msg, uint(msgType), uint(replayId)}}
			js, _ := json.Marshal(m)
			err := c.Conn.WriteMessage(websocket.TextMessage, js)
			if err != nil {
			}
		}
	}
}

// 用户上线
func (c client) wsOnlineClients() {
	wsMutex.Lock()
	pool[c.Uid] = append(pool[c.Uid], &c)
	wsMutex.Unlock()
	//通知前端接收离线消息
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
	return
}
