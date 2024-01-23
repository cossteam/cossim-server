package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/api/model"
	"github.com/cossim/coss-server/interface/msg/config"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/time"
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

// @Summary 发送私聊消息
// @Description 发送私聊消息
// @Accept  json
// @Produce  json
// @param request body model.SendUserMsgRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /msg/send/user [post]
func sendUserMsg(c *gin.Context) {
	req := new(model.SendUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	userRelationStatus1, err := relationClient.GetUserRelation(context.Background(), &relation.GetUserRelationRequest{
		UserId:   thisId,
		FriendId: req.ReceiverId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	if userRelationStatus1.Status != relation.RelationStatus_RELATION_STATUS_ADDED {
		response.SetFail(c, "好友关系不正常", nil)
		return
	}

	userRelationStatus2, err := relationClient.GetUserRelation(context.Background(), &relation.GetUserRelationRequest{
		UserId:   req.ReceiverId,
		FriendId: thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	if userRelationStatus2.Status != relation.RelationStatus_RELATION_STATUS_ADDED {
		response.SetFail(c, "好友关系不正常", nil)
		return
	}

	dialogs, err := dialogClient.GetDialogByIds(context.Background(), &relation.GetDialogByIdsRequest{
		DialogIds: []uint32{req.DialogId},
	})
	if err != nil {
		c.Error(err)
		return
	}
	if len(dialogs.Dialogs) == 0 {
		response.SetFail(c, "对话不存在", nil)
		return
	}
	_, err = dialogClient.GetDialogUserByDialogIDAndUserID(context.Background(), &relation.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	message, err := msgClient.SendUserMessage(context.Background(), &msg.SendUserMsgRequest{
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
			response.SetSuccess(c, "发送成功", nil)
			return
		}
	}
	wsMsg := config.WsMsg{Uid: req.ReceiverId, Event: config.SendUserMessageEvent, SendAt: time.Now(), Data: &model.WsUserMsg{
		SendAt:   time.Now(),
		DialogId: req.DialogId,
		SenderId: thisId,
		Content:  req.Content,
		MsgType:  req.Type,
		ReplayId: req.ReplayId,
	}}

	if enc.IsEnable() {
		marshal, err := json.Marshal(wsMsg)
		if err != nil {
			logger.Error("json解析失败", zap.Error(err))
			return
		}
		message, err := enc.GetSecretMessage(string(marshal), req.ReceiverId)
		if err != nil {
			return
		}
		err = rabbitMQClient.PublishEncryptedMessage(req.ReceiverId, message)
		if err != nil {
			fmt.Println("发布消息失败：", err)
			response.SetFail(c, "发送失败", nil)
			return
		}
		response.SetSuccess(c, "发送成功", nil)
		return
	}
	err = rabbitMQClient.PublishMessage(req.ReceiverId, wsMsg)
	if err != nil {
		fmt.Println("发布消息失败：", err)
		response.SetFail(c, "发送失败", nil)
		return
	}
	response.SetSuccess(c, "发送成功", message)
}

// @Summary 发送群聊消息
// @Description 发送群聊消息
// @Accept  json
// @Produce  json
// @param request body model.SendGroupMsgRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /msg/send/group [post]
func sendGroupMsg(c *gin.Context) {
	req := new(model.SendGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	if req.GroupId == 0 {
		response.SetFail(c, "群聊id不正确", nil)
		return
	}

	groupRelation, err := userGroupClient.GetGroupRelation(context.Background(), &relation.GetGroupRelationRequest{
		GroupId: req.GroupId,
		UserId:  thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	if groupRelation.Status != relation.GroupRelationStatus_GroupStatusJoined {
		response.SetFail(c, "用户在群里中状态不正常", nil)
		return
	}
	if groupRelation.MuteEndTime > time.Now() && groupRelation.MuteEndTime != 0 {
		response.SetFail(c, "用户禁言中", nil)
		return
	}
	dialogs, err := dialogClient.GetDialogByIds(context.Background(), &relation.GetDialogByIdsRequest{
		DialogIds: []uint32{req.DialogId},
	})
	if err != nil {
		c.Error(err)
		return
	}
	if len(dialogs.Dialogs) == 0 {
		response.SetFail(c, "对话不存在", nil)
		return
	}
	_, err = dialogClient.GetDialogUserByDialogIDAndUserID(context.Background(), &relation.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	message, err := msgClient.SendGroupMessage(context.Background(), &msg.SendGroupMsgRequest{
		DialogId: req.DialogId,
		GroupId:  req.GroupId,
		UserId:   thisId,
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
	uids, err := userGroupClient.GetGroupUserIDs(context.Background(), &relation.GroupIDRequest{
		GroupId: req.GroupId,
	})
	sendWsGroupMsg(uids.UserIds, thisId, req.GroupId, req.Content, req.Type, req.ReplayId, req.DialogId)
	response.SetSuccess(c, "发送成功", gin.H{"msg_id": message.MsgId})
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
// @Success		200 {object} model.Response{}
// @Router /msg/list/user [get]
func getUserMsgList(c *gin.Context) {
	var num = c.Query("page_num")
	var size = c.Query("page_size")
	var id = c.Query("user_id")
	var msgType = c.Query("type")
	var content = c.Query("content")

	if num == "" || size == "" || id == "" {
		response.SetFail(c, "参数错误", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	pageNum, _ := strconv.Atoi(num)
	pageSize, _ := strconv.Atoi(size)
	mt, _ := strconv.Atoi(msgType)
	if pageNum == 0 || pageSize == 0 {
		response.SetFail(c, "参数错误", nil)
		return
	}

	var msgListRequest = &model.MsgListRequest{
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
	response.SetSuccess(c, "获取成功", msg)
}

// 获取用户对话列表
// @Summary 获取用户对话列表
// @Description 获取用户对话列表
// @Accept  json
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /msg/dialog/list [get]
func getUserDialogList(c *gin.Context) {
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	//获取对话id
	ids, err := dialogClient.GetUserDialogList(context.Background(), &relation.GetUserDialogListRequest{
		UserId: thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	//获取对话信息
	infos, err := dialogClient.GetDialogByIds(context.Background(), &relation.GetDialogByIdsRequest{
		DialogIds: ids.DialogIds,
	})
	fmt.Println("获取对话信息", zap.Any("infos", infos))
	if err != nil {
		c.Error(err)
		return
	}
	//获取最后一条消息
	dialogIds, err := msgClient.GetLastMsgsByDialogIds(context.Background(), &msg.GetLastMsgsByDialogIdsRequest{
		DialogIds: ids.DialogIds,
	})
	if err != nil {
		logger.Error("获取消息失败", zap.Error(err))
		return
	}
	//封装响应数据
	var responseList = make([]model.UserDialogListResponse, 0)
	for _, v := range infos.Dialogs {
		fmt.Println("获取最后一条消息", zap.Any("v", v))
		var re model.UserDialogListResponse
		//用户
		if v.Type == 0 {
			users, _ := dialogClient.GetDialogUsersByDialogID(context.Background(), &relation.GetDialogUsersByDialogIDRequest{
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
			//re.UserId = v.OwnerId
			re.GroupId = info.Id
			re.DialogId = v.Id
		}
		// 匹配最后一条消息
		for _, msg := range dialogIds.LastMsgs {
			if msg.DialogId == v.Id {
				re.LastMessage = model.Message{
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
	response.SetSuccess(c, "获取成功", responseList)
}

// @Summary 编辑用户消息
// @Description 编辑用户消息
// @Accept  json
// @Produce  json
// @param request body model.EditUserMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/edit/user [post]
func editUserMsg(c *gin.Context) {
	req := new(model.EditUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	//获取消息
	msginfo, err := msgClient.GetUserMessageById(context.Background(), &msg.GetUserMsgByIDRequest{
		MsgId: req.MsgId,
	})
	if err != nil {
		logger.Error("获取消息失败", zap.Error(err))
		c.Error(err)
		return
	}
	if msginfo.SenderId != thisId {
		response.SetFail(c, "不是你发送的消息", nil)
		return
	}
	// 调用相应的 gRPC 客户端方法来编辑用户消息
	_, err = msgClient.EditUserMessage(context.Background(), &msg.EditUserMsgRequest{
		UserMessage: &msg.UserMessage{
			Id:      req.MsgId,
			Content: req.Content,
		},
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", gin.H{"msg_id": req.MsgId})
}

// @Summary 编辑群消息
// @Description 编辑群消息
// @Accept  json
// @Produce  json
// @param request body model.EditGroupMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/edit/group [post]
func editGroupMsg(c *gin.Context) {
	req := new(model.EditGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	//获取消息
	msginfo, err := msgClient.GetGroupMessageById(context.Background(), &msg.GetGroupMsgByIDRequest{
		MsgId: req.MsgId,
	})
	if err != nil {
		logger.Error("获取消息失败", zap.Error(err))
		c.Error(err)
		return
	}
	if msginfo.UserId != thisId {
		response.SetFail(c, "不是你发送的消息", nil)
		return
	}
	// 调用相应的 gRPC 客户端方法来编辑群消息
	_, err = msgClient.EditGroupMessage(context.Background(), &msg.EditGroupMsgRequest{
		GroupMessage: &msg.GroupMessage{
			Id:      req.MsgId,
			Content: req.Content,
		},
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", gin.H{"msg_id": req.MsgId})
}

// @Summary 撤回用户消息
// @Description 撤回用户消息
// @Accept  json
// @Produce  json
// @param request body model.RecallUserMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/recall/user [post]
func recallUserMsg(c *gin.Context) {
	req := new(model.RecallUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	//获取消息
	msginfo, err := msgClient.GetUserMessageById(context.Background(), &msg.GetUserMsgByIDRequest{
		MsgId: req.MsgId,
	})
	if err != nil {
		logger.Error("获取消息失败", zap.Error(err))
		c.Error(err)
		return
	}
	if msginfo.SenderId != thisId {
		response.SetFail(c, "不是你发送的消息", nil)
		return
	}
	//判断发送时间是否超过两分钟
	if time.Now()-msginfo.CreatedAt > 120 {
		response.SetFail(c, "该条消息发送时间已经超过两分钟，不能撤回", nil)
		return
	}

	// 调用相应的 gRPC 客户端方法来撤回用户消息
	msg, err := msgClient.DeleteUserMessage(context.Background(), &msg.DeleteUserMsgRequest{
		MsgId: req.MsgId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", gin.H{"msg_id": msg.Id})
}

// @Summary 撤回群消息
// @Description 撤回群消息
// @Accept  json
// @Produce  json
// @param request body model.RecallGroupMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/recall/group [post]
func recallGroupMsg(c *gin.Context) {
	req := new(model.RecallGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	//获取消息
	msginfo, err := msgClient.GetGroupMessageById(context.Background(), &msg.GetGroupMsgByIDRequest{
		MsgId: req.MsgId,
	})
	if err != nil {
		logger.Error("获取消息失败", zap.Error(err))
		c.Error(err)
		return
	}
	if msginfo.UserId != thisId {
		response.SetFail(c, "不是你发送的消息", nil)
		return
	}
	//判断发送时间是否超过两分钟
	if time.Now()-msginfo.CreatedAt > 120 {
		response.SetFail(c, "该条消息发送时间已经超过两分钟，不能撤回", nil)
		return
	}

	// 调用相应的 gRPC 客户端方法来撤回群消息
	msg, err := msgClient.DeleteGroupMessage(context.Background(), &msg.DeleteGroupMsgRequest{
		MsgId: req.MsgId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", gin.H{"msg_id": msg.Id})
}

// @Summary 标注用户消息状态
// @Description 标注用户消息状态
// @Accept  json
// @Produce  json
// @param request body model.LabelUserMessageRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/label/user [post]
func labelUserMessage(c *gin.Context) {
	req := new(model.LabelUserMessageRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if !model.IsValidLabelMsgType(req.IsLabel) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	// 获取用户消息
	msginfo, err := msgClient.GetUserMessageById(context.Background(), &msg.GetUserMsgByIDRequest{
		MsgId: req.MsgID,
	})
	if err != nil {
		logger.Error("获取用户消息失败", zap.Error(err))
		c.Error(err)
		return
	}
	//判断是否在对话内
	userIds, err := dialogClient.GetDialogUsersByDialogID(context.Background(), &relation.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	found := false
	for _, v := range userIds.UserIds {
		if v == thisId {
			found = true
			break
		}
	}
	if !found {
		response.SetFail(c, "不在对话内", nil)
		return
	}

	// 调用 gRPC 客户端方法将用户消息设置为标注状态
	_, err = msgClient.SetUserMsgLabel(context.Background(), &msg.SetUserMsgLabelRequest{
		MsgId:   req.MsgID,
		IsLabel: msg.MsgLabel(req.IsLabel),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "用户消息标注成功", nil)
}

// @Summary 标注群聊消息状态
// @Description 标注群聊消息状态
// @Accept  json
// @Produce  json
// @param request body model.LabelGroupMessageRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/label/group [post]
func labelGroupMessage(c *gin.Context) {
	req := new(model.LabelGroupMessageRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if !model.IsValidLabelMsgType(req.IsLabel) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	// 获取群聊消息
	msginfo, err := msgClient.GetGroupMessageById(context.Background(), &msg.GetGroupMsgByIDRequest{
		MsgId: req.MsgID,
	})
	if err != nil {
		logger.Error("获取群聊消息失败", zap.Error(err))
		c.Error(err)
		return
	}

	//判断是否在对话内
	userIds, err := dialogClient.GetDialogUsersByDialogID(context.Background(), &relation.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	found := false
	for _, v := range userIds.UserIds {
		if v == thisId {
			found = true
			break
		}
	}
	if !found {
		response.SetFail(c, "不在对话内", nil)
		return
	}

	// 调用 gRPC 客户端方法将群聊消息设置为标注状态
	_, err = msgClient.SetGroupMsgLabel(context.Background(), &msg.SetGroupMsgLabelRequest{
		MsgId:   req.MsgID,
		IsLabel: msg.MsgLabel(req.IsLabel),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "群聊消息标注成功", nil)
}

// 获取私聊标注信息
// @Summary 获取私聊标注信息
// @Description 获取私聊标注信息
// @Accept  json
// @Produce  json
// @Param dialog_id query string true "对话id"
// @Success		200 {object} model.Response{}
// @Router /msg/label/user [get]
func getUserLabelMsgList(c *gin.Context) {
	var id = c.Query("dialog_id")

	if id == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	dialogId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = dialogClient.GetDialogUserByDialogIDAndUserID(context.Background(), &relation.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   thisId,
		DialogId: uint32(dialogId),
	})
	if err != nil {
		c.Error(err)
		return
	}

	msgs, err := msgClient.GetUserMsgLabelByDialogId(context.Background(), &msg.GetUserMsgLabelByDialogIdRequest{
		DialogId: uint32(dialogId),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", msgs)
}

// 获取群聊标注信息
// @Summary 获取群聊标注信息
// @Description 获取群聊标注信息
// @Accept  json
// @Produce  json
// @Param dialog_id query string true "对话id"
// @Success		200 {object} model.Response{}
// @Router /msg/label/group [get]
func getGroupLabelMsgList(c *gin.Context) {
	var id = c.Query("dialog_id")

	if id == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	dialogId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = dialogClient.GetDialogUserByDialogIDAndUserID(context.Background(), &relation.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   thisId,
		DialogId: uint32(dialogId),
	})
	if err != nil {
		c.Error(err)
		return
	}

	msgs, err := msgClient.GetGroupMsgLabelByDialogId(context.Background(), &msg.GetGroupMsgLabelByDialogIdRequest{
		DialogId: uint32(dialogId),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", msgs)
}

// 推送私聊消息
func sendWsUserMsg(senderId, receiverId string, msg string, msgType uint, replayId uint, dialogId uint32) {
	//遍历该用户所有客户端
	for _, c := range Pool[receiverId] {
		m := config.WsMsg{Uid: receiverId, Event: config.SendUserMessageEvent, Rid: c.Rid, SendAt: time.Now(), Data: &model.WsUserMsg{SenderId: senderId, Content: msg, MsgType: msgType, ReplayId: replayId, SendAt: time.Now(), DialogId: dialogId}}
		js, _ := json.Marshal(m)
		message, err := enc.GetSecretMessage(string(js), receiverId)
		if err != nil {
			return
		}
		err = c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			logger.Error("send msg err", zap.Error(err))
			return
		}
	}
}

// 推送群聊消息
func sendWsGroupMsg(uIds []string, userId string, groupId uint32, msg string, msgType uint32, replayId uint32, dialogId uint32) {
	//发送群聊消息
	for _, uid := range uIds {
		//在线则推送ws
		if _, ok := Pool[uid]; ok {
			for _, c := range Pool[uid] {
				m := config.WsMsg{Uid: uid, Event: config.SendGroupMessageEvent, Rid: c.Rid, Data: &model.WsGroupMsg{GroupId: int64(groupId), UserId: userId, Content: msg, MsgType: uint(msgType), ReplayId: uint(replayId), SendAt: time.Now(), DialogId: dialogId}}
				js, _ := json.Marshal(m)
				message, err := enc.GetSecretMessage(string(js), uid)
				if err != nil {
					return
				}
				fmt.Println("推送ws消息给", uid, ":", message)
				err = c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					fmt.Println("send ws msg err:", err)
					continue
				}
			}
		}
		//否则推送到消息队列
		msg := config.WsMsg{Uid: uid, Event: config.SendGroupMessageEvent, SendAt: time.Now(), Data: &model.WsGroupMsg{
			GroupId:  int64(groupId),
			UserId:   userId,
			Content:  msg,
			MsgType:  uint(msgType),
			ReplayId: uint(replayId),
			SendAt:   time.Now(),
			DialogId: dialogId,
		}}
		if enc.IsEnable() {
			marshal, err := json.Marshal(msg)
			if err != nil {
				logger.Error("json解析失败", zap.Error(err))
				return
			}
			message, err := enc.GetSecretMessage(string(marshal), uid)
			if err != nil {
				return
			}
			err = rabbitMQClient.PublishEncryptedMessage(uid, message)
			if err != nil {
				fmt.Println("发布消息失败：", err)
				return
			}
			return
		}
		err := rabbitMQClient.PublishMessage(uid, msg)
		if err != nil {
			fmt.Println("发布消息失败：", err)
			return
		}
	}
}

// SendMsg 推送消息
func SendMsg(uid string, event config.WSEventType, data interface{}) {
	m := config.WsMsg{Uid: uid, Event: event, Rid: 0, Data: data, SendAt: time.Now()}
	if _, ok := Pool[uid]; !ok {
		//不在线则推送到消息队列
		err := rabbitMQClient.PublishMessage(uid, m)
		if err != nil {
			fmt.Println("发布消息失败：", err)
			return
		}
		return
	}
	for _, c := range Pool[uid] {
		m.Rid = c.Rid
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
	msg := config.WsMsg{Uid: c.Uid, Event: config.OnlineEvent, Rid: c.Rid, SendAt: time.Now()}
	js, _ := json.Marshal(msg)
	if enc == nil {
		logger.Error("加密客户端错误", zap.Error(nil))
		return
	}
	message, err := enc.GetSecretMessage(string(js), c.Uid)
	if err != nil {
		fmt.Println("加密失败：", err)
		return
	}
	//上线推送消息
	c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	for {
		msg, ok, err := msg_queue.ConsumeMessages(c.Uid, c.queue)
		if err != nil || !ok {
			//c.queue.Close()
			//拉取完之后删除队列
			_ = rabbitMQClient.DeleteEmptyQueue(c.Uid)
			return
		}
		c.Conn.WriteMessage(websocket.TextMessage, msg.Body)
	}
}

// 用户离线
func (c client) wsOfflineClients() {
	wsMutex.Lock()
	defer wsMutex.Unlock()
	// 删除用户
	delete(Pool, c.Uid)
}
