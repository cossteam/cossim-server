package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/config"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	msg "github.com/cossim/coss-server/service/msg/api/v1"
	relation "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	wsRid    int64 = 0            //全局客户端id
	wsMutex        = sync.Mutex{} //锁
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	Pool = make(map[string][]*client)
)

type client struct {
	Conn  *websocket.Conn
	Uid   string //客户端所有者
	Rid   int64  //客户端id
	queue *amqp091.Channel
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
	messages := rabbitMQClient.GetChannel()
	if messages.IsClosed() {
		log.Fatal("Channel is Closed")
	}
	client := &client{
		Conn:  conn,
		Uid:   "",
		Rid:   wsRid,
		queue: messages,
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
	DialogId   uint32 `json:"dialog_id" binding:"required"`
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
		DialogId:   req.DialogId,
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
	if _, ok := Pool[req.ReceiverId]; ok {
		if len(Pool[req.ReceiverId]) > 0 {
			sendWsUserMsg(thisId, req.ReceiverId, req.Content, req.Type, req.ReplayId, req.DialogId)
			response.Success(c, "发送成功", nil)
			return
		}
	}
	msg := config.WsMsg{Uid: req.ReceiverId, Event: config.SendUserMessageEvent, SendAt: time.Now().Unix(), Data: &wsUserMsg{
		SendAt:   time.Now().Unix(),
		DialogId: req.DialogId,
		SenderId: thisId,
		Content:  req.Content,
		MsgType:  req.Type,
		ReplayId: req.ReplayId,
	}}
	// todo 记录离线推送
	err = rabbitMQClient.PublishMessage(req.ReceiverId, msg)
	if err != nil {
		fmt.Println("发布消息失败：", err)
		response.Fail(c, "发送好友请求失败", nil)
		return
	}
	response.Success(c, "发送成功", nil)
}

type SendGroupMsgRequest struct {
	DialogId uint32 `json:"dialog_id" binding:"required"`
	GroupId  uint32 `json:"group_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     uint32 `json:"type" binding:"required"`
	ReplayId uint32 `json:"replay_id" binding:"required"`
	SendAt   int64  `json:"send_at" binding:"required"`
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
		DialogId: req.DialogId,
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
		GroupId: req.GroupId,
	})
	sendWsGroupMsg(uids.UserIds, thisId, req.GroupId, req.Content, req.Type, req.ReplayId, req.DialogId)
	response.Success(c, "发送成功", nil)
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
	response.Success(c, "获取成功", msg)
}

type ConversationType uint

const (
	UserConversation ConversationType = iota
	GroupConversation
)

type UserDialogListResponse struct {
	DialogId uint32 `json:"dialog_id"`
	UserId   string `json:"user_id,omitempty"`
	GroupId  uint32 `json:"group_id,omitempty"`
	// 会话类型
	DialogType ConversationType `json:"dialog_type"`
	// 会话名称
	DialogName string `json:"dialog_name"`
	// 会话头像
	DialogAvatar string `json:"dialog_avatar"`
	// 会话未读消息数
	DialogUnreadCount int     `json:"dialog_unread_count"`
	LastMessage       Message `json:"last_message"`
}
type Message struct {
	// 消息类型
	MsgType uint `json:"msg_type"`
	// 消息内容
	Content string `json:"content"`
	// 消息发送者
	SenderId string `json:"sender_id"`
	// 消息发送时间
	SendTime int64 `json:"send_time"`
	// 消息id
	MsgId uint64 `json:"msg_id"`
}

// 获取用户对话列表
// @Summary 获取用户对话列表
// @Description 获取用户对话列表
// @Accept  json
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /msg/dialog/list [get]
func getUserDialogList(c *gin.Context) {
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.Fail(c, err.Error(), nil)
		return
	}
	//获取对话id
	ids, err := dialogClient.GetUserDialogList(context.Background(), &msg.GetUserDialogListRequest{
		UserId: thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	//获取对话信息
	infos, err := dialogClient.GetDialogByIds(context.Background(), &msg.GetDialogByIdsRequest{
		DialogIds: ids.DialogIds,
	})
	//获取最后一条消息
	dialogIds, err := msgClient.GetLastMsgsByDialogIds(context.Background(), &msg.GetLastMsgsByDialogIdsRequest{
		DialogIds: ids.DialogIds,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//封装响应数据
	var responseList = make([]UserDialogListResponse, 0)
	for _, v := range infos.Dialogs {
		var re UserDialogListResponse
		//用户
		if v.Type == 0 {
			users, _ := dialogClient.GetDialogUsersByDialogID(context.Background(), &msg.GetDialogUsersByDialogIDRequest{
				DialogId: v.Id,
			})
			if len(users.UserIds) == 0 {
				continue
			}
			for _, id := range users.UserIds {
				if id == thisId {
					continue
				}
				info, err := userClient.UserInfo(context.Background(), &user.UserInfoRequest{
					UserId: id,
				})
				if err != nil {
					fmt.Println(err)
					continue
				}
				re.DialogId = v.Id
				re.DialogAvatar = info.Avatar
				re.DialogName = info.NickName
				re.DialogType = 0
				re.DialogUnreadCount = 10
				re.UserId = info.UserId
				break
			}

		} else if v.Type == 1 {
			//群聊
			info, err := groupClient.GetGroupInfoByGid(context.Background(), &groupApi.GetGroupInfoRequest{
				Gid: v.GroupId,
			})
			if err != nil {
				fmt.Println(err)
				continue
			}
			re.DialogAvatar = info.Avatar
			re.DialogName = info.Name
			re.DialogType = 1
			re.DialogUnreadCount = 10
			re.UserId = v.OwnerId
			re.DialogId = v.Id
		}
		// 匹配最后一条消息
		for _, msg := range dialogIds.LastMsgs {
			if msg.DialogId == v.Id {
				re.LastMessage = Message{
					MsgId:    uint64(msg.Id),
					Content:  msg.Content,
					SenderId: msg.SenderId,
					SendTime: msg.CreatedAt,
					MsgType:  uint(msg.Type),
				}
				break
			}
		}

		responseList = append(responseList, re)
	}
	//根据发送时间排序
	sort.Slice(responseList, func(i, j int) bool {
		return responseList[i].LastMessage.SendTime > responseList[j].LastMessage.SendTime
	})
	response.Success(c, "获取成功", responseList)
}

type wsUserMsg struct {
	SenderId string `json:"sender_id"`
	Content  string `json:"content"`
	MsgType  uint   `json:"msgType"`
	ReplayId uint   `json:"reply_id"`
	SendAt   int64  `json:"send_at"`
	DialogId uint32 `json:"dialog_id"`
}

// 推送私聊消息
func sendWsUserMsg(senderId, receiverId string, msg string, msgType uint, replayId uint, dialogId uint32) {
	//遍历该用户所有客户端
	for _, c := range Pool[receiverId] {
		m := config.WsMsg{Uid: receiverId, Event: config.SendUserMessageEvent, Rid: c.Rid, SendAt: time.Now().Unix(), Data: &wsUserMsg{senderId, msg, msgType, replayId, time.Now().Unix(), dialogId}}
		js, _ := json.Marshal(m)
		err := c.Conn.WriteMessage(websocket.TextMessage, js)
		if err != nil {
			logger.Error("send msg err", zap.Error(err))
			return
		}
	}
}

type wsGroupMsg struct {
	GroupId  int64  `json:"group_id"`
	UserId   string `json:"uid"`
	Content  string `json:"content"`
	MsgType  uint   `json:"msgType"`
	ReplayId uint   `json:"reply_id"`
	SendAt   int64  `json:"send_at"`
	DialogId uint32 `json:"dialog_id"`
}

// 推送群聊消息
func sendWsGroupMsg(uIds []string, userId string, groupId uint32, msg string, msgType uint32, replayId uint32, dialogId uint32) {
	//发送群聊消息
	for _, uid := range uIds {
		////遍历该用户所有客户端
		//if uid == userId {
		//	continue
		//}
		for _, c := range Pool[uid] {
			m := config.WsMsg{Uid: uid, Event: config.SendGroupMessageEvent, Rid: c.Rid, Data: &wsGroupMsg{int64(groupId), userId, msg, uint(msgType), uint(replayId), time.Now().Unix(), dialogId}}
			js, _ := json.Marshal(m)
			err := c.Conn.WriteMessage(websocket.TextMessage, js)
			if err != nil {
			}
		}
		msg := config.WsMsg{Uid: uid, Event: config.SendGroupMessageEvent, SendAt: time.Now().Unix(), Data: &wsGroupMsg{
			GroupId:  int64(groupId),
			UserId:   userId,
			Content:  msg,
			MsgType:  uint(msgType),
			ReplayId: uint(replayId),
			SendAt:   time.Now().Unix(),
			DialogId: dialogId,
		}}
		// todo 记录离线推送
		err := rabbitMQClient.PublishMessage(uid, msg)
		if err != nil {
			fmt.Println("发布消息失败：", err)
		}
	}
}

// SendMsg 推送消息
func SendMsg(uid string, event config.WSEventType, data interface{}) {
	if _, ok := Pool[uid]; !ok {
		return
	}
	for _, c := range Pool[uid] {
		m := config.WsMsg{Uid: uid, Event: event, Rid: c.Rid, Data: data, SendAt: time.Now().Unix()}
		js, _ := json.Marshal(m)
		err := c.Conn.WriteMessage(websocket.TextMessage, js)
		if err != nil {
			logger.Error("send msg err", zap.Error(err))
			return
		}
	}
}

// 推送多个用户消息
func SendMsgToUsers(uids []string, event config.WSEventType, data interface{}) {
	for _, uid := range uids {
		SendMsg(uid, event, data)
	}
}

// 用户上线
func (c client) wsOnlineClients() {
	wsMutex.Lock()
	Pool[c.Uid] = append(Pool[c.Uid], &c)
	wsMutex.Unlock()
	//通知前端接收离线消息
	msg := config.WsMsg{Uid: c.Uid, Event: config.OnlineEvent, Rid: c.Rid, SendAt: time.Now().Unix()}
	js, _ := json.Marshal(msg)
	//上线推送消息
	c.Conn.WriteMessage(websocket.TextMessage, js)
	//go func() {
	for {
		msg, ok, err := msg_queue.ConsumeMessages(c.Uid, c.queue)
		if err != nil || !ok {
			//c.queue.Close()
			return
		}
		c.Conn.WriteMessage(websocket.TextMessage, msg.Body)
	}
	//}()
}

// 用户离线
func (c client) wsOfflineClients() {
	wsMutex.Lock()
	defer wsMutex.Unlock()
	// 删除用户
	delete(Pool, c.Uid)
}
