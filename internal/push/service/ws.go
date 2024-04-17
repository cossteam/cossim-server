package service

import (
	"context"
	"encoding/json"
	"fmt"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/internal/push/api/http/model"
	"github.com/cossim/coss-server/internal/push/connect"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils"
	myos "github.com/cossim/coss-server/pkg/utils/time"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func (s *Service) Ws(ctx context.Context, conn *websocket.Conn, uid string, driverId string, deviceType, token string) {
	//新建websocket客户端
	wsRid++
	cli := connect.NewWebsocketClient(wsRid, conn)
	bucket := s.Buckets[constants.DriverType(deviceType)]

	if bucket.GetUserIndex(uid) == nil {
		bucket.Push(connect.NewUserIndex(uid))
	}

	index := bucket.GetUserIndex(uid)

	defer conn.Close()
	defer func(client *connect.WebsocketClient, index *connect.UserIndex, rid int64) {
		client.Conn.Close()
		index.DeleteByRid(rid)
	}(cli, index, wsRid)

	//设备限制
	if s.ac.MultipleDeviceLimit.Enable {
		if index.GetLength() >= s.ac.MultipleDeviceLimit.Max {
			s.logger.Error("用户登录设备达到上限")
			return
		}
	}

	index.Push(cli)
	bytes, err := utils.StructToBytes(model.OnlineEventData{DriverType: deviceType})
	if err != nil {
		s.logger.Error("序列化失败", zap.Error(err))
		return
	}
	//保存到线程池
	s.wsOnlineClients(ctx, &pushgrpcv1.WsMsg{Uid: uid, DriverId: driverId, Event: pushgrpcv1.WSEventType_OnlineEvent, Rid: wsRid, SendAt: pkgtime.Now(), Data: &any.Any{Value: bytes}}, cli)

	go cli.Conn.CheckHeartbeat(30 * time.Second)

	//更新登录信息
	keys, err := s.redisClient.ScanKeys(uid + ":" + deviceType + ":*")
	if err != nil {
		s.logger.Error("获取用户信息失败1", zap.Error(err))
		return
	}

	for _, key := range keys {
		v, err := s.redisClient.GetKey(key)
		if err != nil {
			s.logger.Error("获取用户信息失败", zap.Error(err))
			return
		}
		strKey := v.(string)
		info, err := cache.GetUserInfo(strKey)
		if err != nil {
			s.logger.Error("获取用户信息失败", zap.Error(err))
			return
		}
		if info.Token == token {
			info.Rid = cli.Rid
			resp := cache.GetUserInfoToInterfaces(info)
			err := s.redisClient.SetKey(key, resp, 60*60*24*7*time.Second)
			if err != nil {
				s.logger.Error("保存用户信息失败", zap.Error(err))
				return
			}
			break
		}
	}
	//读取客户端消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.logger.Error("读取消息失败", zap.Error(err))
			//删除redis里的rid
			keys, err := s.redisClient.ScanKeys(uid + ":" + deviceType + ":*")
			if err != nil {
				s.logger.Error("获取用户信息失败1", zap.Error(err))
				return
			}

			for _, key := range keys {
				v, err := s.redisClient.GetKey(key)
				if err != nil {
					s.logger.Error("获取用户信息失败", zap.Error(err))
					return
				}
				strKey := v.(string)
				info, err := cache.GetUserInfo(strKey)
				if err != nil {
					s.logger.Error("获取用户信息失败", zap.Error(err))
					return
				}
				if info.Token == token {
					info.Rid = 0
					resp := cache.GetUserInfoToInterfaces(info)
					err := s.redisClient.SetKey(key, resp, 60*60*24*7*time.Second)
					if err != nil {
						s.logger.Error("保存用户信息失败", zap.Error(err))
						return
					}
					break
				}
			}
			//用户下线
			s.wsOfflineClients(ctx, uid)
			err = cli.Conn.Close()
			if err != nil {
				return
			}
			index.DeleteByRid(wsRid)
			if index.GetLength() == 0 {
				s.logger.Info("该用户最后一个客户端已经离线,删除索引")
				bucket.DeleteByUserID(uid)
			}
			return
		}
	}
}

func (s *Service) PushWs(ctx context.Context, msg *pushgrpcv1.WsMsg) (*pushgrpcv1.PushResponse, error) {
	resp := &pushgrpcv1.PushResponse{}
	pushd := false

	bytes, err := wsMsgToJSON(msg, false)
	if err != nil {
		return nil, err
	}
	message, err := s.enc.GetSecretMessage(ctx, string(bytes), msg.Uid)
	if err != nil {
		return nil, err
	}

	for _, i2 := range s.Buckets {
		ui := i2.GetUserIndex(msg.Uid)
		if !msg.PushOffline && ui == nil {
			continue
		}

		if ui != nil && ui.GetLength() > 0 {
			err = i2.SendMessage(msg.Uid, message)
			if err != nil {
				return resp, err
			}
			pushd = true
		}

	}
	if msg.PushOffline && !pushd {
		//不在线则推送到消息队列
		err := s.rabbitMQClient.PublishMessage(msg.Uid, message)
		if err != nil {
			s.logger.Error("发布消息失败", zap.Error(err))
		}
	}

	return resp, nil
}

func (s *Service) PushWsBatch(ctx context.Context, request *pushgrpcv1.PushWsBatchRequest) (*pushgrpcv1.PushResponse, error) {
	resp := &pushgrpcv1.PushResponse{}
	for _, i2 := range s.Buckets {
		for _, msg := range request.Msgs {
			bytes, err := wsMsgToJSON(msg, false)
			if err != nil {
				return nil, err
			}

			message, err := s.enc.GetSecretMessage(ctx, string(bytes), msg.Uid)
			if err != nil {
				return nil, err
			}

			ui := i2.GetUserIndex(msg.Uid)

			if !msg.PushOffline && ui == nil {
				continue
			}
			if msg.PushOffline && ui == nil {
				//不在线则推送到消息队列
				err := s.rabbitMQClient.PublishMessage(msg.Uid, message)
				if err != nil {
					s.logger.Error("发布消息失败：", zap.Error(err))
					continue
				}
				continue
			}

			err = i2.SendMessage(msg.Uid, message)
			if err != nil {
				return resp, err
			}
		}
	}
	return resp, nil
}

func (s *Service) PushWsBatchByUserIds(ctx context.Context, request *pushgrpcv1.PushWsBatchByUserIdsRequest) (*pushgrpcv1.PushResponse, error) {
	resp := &pushgrpcv1.PushResponse{}
	for _, id := range request.UserIds {
		msg := &pushgrpcv1.WsMsg{
			Uid:         id,
			Event:       request.Event,
			Rid:         0,
			DriverId:    request.DriverId,
			SendAt:      myos.Now(),
			Data:        request.Data,
			PushOffline: request.PushOffline,
		}

		//bytes, err := utils.StructToBytes(msg)
		//if err != nil {
		//	return resp, err
		//}
		bytes, err := wsMsgToJSON(msg, false)
		if err != nil {
			return nil, err
		}

		if s.enc == nil {
			s.logger.Error("加密客户端错误", zap.Error(nil))
			return nil, fmt.Errorf("加密客户端错误%v", zap.Error(nil))
		}

		message, err := s.enc.GetSecretMessage(ctx, string(bytes), msg.Uid)
		if err != nil {
			return nil, err
		}

		for _, bucket := range s.Buckets {
			ui := bucket.GetUserIndex(msg.Uid)

			if !msg.PushOffline && ui == nil {
				continue
			}
			if msg.PushOffline && ui == nil {
				//不在线则推送到消息队列
				err := s.rabbitMQClient.PublishMessage(msg.Uid, message)
				if err != nil {
					s.logger.Error("发布消息失败：", zap.Error(err))
					continue
				}
				continue
			}

			err := bucket.SendMessage(id, string(message))
			if err != nil {
				return resp, err
			}
		}
	}
	return resp, nil
}

// 用户上线
func (s *Service) wsOnlineClients(ctx context.Context, msg *pushgrpcv1.WsMsg, c *connect.WebsocketClient) {

	js, err := wsMsgToJSON(msg, false)
	if err != nil {
		s.logger.Error("上线失败：", zap.Error(err))
		return
	}
	if s.enc == nil {
		s.logger.Error("加密客户端错误", zap.Error(nil))
		return
	}

	message, err := s.enc.GetSecretMessage(ctx, string(js), msg.Uid)
	if err != nil {
		s.logger.Error("上线失败：", zap.Error(err))
		return
	}

	//上线推送消息
	err = c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		s.logger.Error("上线失败：", zap.Error(err))
		return
	}

	err = s.pushAllFriendOnlineStatus(ctx, c, msg.Uid)
	if err != nil {
		s.logger.Error("上线失败：", zap.Error(err))
		return
	}

	//修改在线状态
	err = s.addUserWsCount(ctx, msg.Uid)
	if err != nil {
		s.logger.Error("上线失败：", zap.Error(err))
		return
	}

	for {
		msg2, ok, err := msg_queue.ConsumeMessages(msg.Uid, s.rabbitMQClient.GetChannel())
		if err != nil || !ok {
			//c.queue.Stop()
			//拉取完之后删除队列
			_ = s.rabbitMQClient.DeleteEmptyQueue(msg.Uid)
			return
		}

		var a interface{}
		err = json.Unmarshal(msg2, &a)
		if err != nil {
			s.logger.Error("上线失败：", zap.Error(err))
			return
		}

		mm := a.(string)
		// 尝试解析消息
		var data2 map[string]interface{}
		err = json.Unmarshal([]byte(mm), &data2)
		if err != nil {
			s.logger.Error("转换消息失败1", zap.Error(err))
			return
		}

		// 尝试将解析后的数据转换为字节
		wsData, err := json.Marshal(data2)
		if err != nil {
			s.logger.Error("转换消息失败2", zap.Error(err))
			return
		}

		// 尝试将数据写入 WebSocket 连接
		err = c.Conn.WriteMessage(websocket.TextMessage, wsData)
		if err != nil {
			s.logger.Error("推送队列离线消息失败", zap.Error(err))
			return
		}
	}
}

func (s *Service) wsOfflineClients(ctx context.Context, uid string) {
	err := s.reduceUserWsCount(ctx, uid)
	if err != nil {
		s.logger.Error("修改在线状态失败：", zap.Error(err))
		return
	}
}

func (s *Service) addUserWsCount(ctx context.Context, uid string) error {

	exists, err := s.redisClient.ExistsKey(uid)
	if err != nil {
		return err
	}

	//给好友推送上线
	err = s.pushFriendStatus(ctx, onlineEvent, uid)
	if err != nil {
		return err
	}

	if !exists {
		return s.redisClient.SetKey(uid, 1, 0)
	} else {
		value, err := s.redisClient.GetKey(uid)
		if err != nil {
			return err
		}
		str := value.(string)
		num, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		num++
		return s.redisClient.SetKey(uid, num, 0)
	}
}

func (s *Service) reduceUserWsCount(ctx context.Context, uid string) error {

	exists, err := s.redisClient.ExistsKey(uid)
	if err != nil {
		return err
	}
	if !exists {
		//给好友推送下线
		err := s.pushFriendStatus(ctx, offlineEvent, uid)
		if err != nil {
			return err
		}
		return nil
	} else {
		value, err := s.redisClient.GetKey(uid)
		if err != nil {
			return err
		}
		str := value.(string)
		num, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		if num == 1 {
			//给好友推送下线
			err := s.pushFriendStatus(ctx, offlineEvent, uid)
			if err != nil {
				return err
			}
			return s.redisClient.DelKey(uid)
		} else {
			num--
			return s.redisClient.SetKey(uid, num, 0)
		}
	}
}

type status uint

const (
	onlineEvent status = iota + 1
	// OfflineEvent 下线事件
	offlineEvent
)

// 给好友推送离线或上线通知
func (s *Service) pushFriendStatus(ctx context.Context, v status, uid string) error {
	//查询所有好友
	list, err := s.relationService.GetFriendList(context.Background(), &relationgrpcv1.GetFriendListRequest{UserId: uid})
	if err != nil {
		return err
	}
	if len(list.FriendList) > 0 {
		bytes, err := utils.StructToBytes(model.FriendOnlineStatusMsg{Status: int32(v), UserId: uid})
		if err != nil {
			return err
		}
		for _, friend := range list.FriendList {
			msg := &pushgrpcv1.WsMsg{Uid: friend.UserId, Event: pushgrpcv1.WSEventType_FriendUpdateOnlineStatusEvent, Rid: 0, SendAt: pkgtime.Now(), Data: &any.Any{Value: bytes}}
			js, _ := wsMsgToJSON(msg, false)
			if s.enc == nil {
				s.logger.Error("推送上线通知失败：", zap.Error(err))
				continue
			}

			if s.enc == nil {
				return fmt.Errorf("加密客户端错误%v", zap.Error(nil))
			}

			message, err := s.enc.GetSecretMessage(ctx, string(js), uid)
			if err != nil {
				s.logger.Error("推送上线通知失败：", zap.Error(err))
				continue
			}

			for _, i2 := range s.Buckets {
				err := i2.SendMessage(friend.UserId, message)
				if err != nil {
					s.logger.Error("推送消息失败", zap.Error(err))
					continue
				}
			}
		}
	}
	return nil
}

// 获取所有好友在线状态
func (s *Service) pushAllFriendOnlineStatus(ctx context.Context, c *connect.WebsocketClient, uid string) error {
	if c.Conn.IsNil() || c.Conn.WriteMessage(websocket.PingMessage, nil) != nil {
		return nil
	}
	//查询所有好友
	list, err := s.relationService.GetFriendList(context.Background(), &relationgrpcv1.GetFriendListRequest{UserId: uid})
	if err != nil {
		return err
	}
	var friendList []model.FriendOnlineStatusMsg

	if len(list.FriendList) > 0 {
		for _, friend := range list.FriendList {
			exists, err := s.redisClient.ExistsKey(friend.UserId)
			if err != nil {
				return err
			}
			if exists {
				value, err := s.redisClient.GetKey(friend.UserId)
				if err != nil {
					return err
				}
				str := value.(string)
				num, err := strconv.Atoi(str)
				if err != nil {
					return err
				}
				if num > 0 {
					friendList = append(friendList, model.FriendOnlineStatusMsg{Status: int32(onlineEvent), UserId: friend.UserId})
				} else {
					friendList = append(friendList, model.FriendOnlineStatusMsg{Status: int32(offlineEvent), UserId: friend.UserId})
				}
			} else {
				friendList = append(friendList, model.FriendOnlineStatusMsg{Status: int32(onlineEvent), UserId: friend.UserId})
			}
		}
	}

	bytes, err := utils.StructToBytes(friendList)
	if err != nil {
		return err
	}

	msg := &pushgrpcv1.WsMsg{Uid: uid, Event: pushgrpcv1.WSEventType_PushAllFriendOnlineStatusEvent, Rid: c.Rid, SendAt: pkgtime.Now(), Data: &any.Any{Value: bytes}}
	js, _ := wsMsgToJSON(msg, true)
	if s.enc == nil {
		s.logger.Error("转换失败：", zap.Error(err))
		return nil
	}
	message, err := s.enc.GetSecretMessage(ctx, string(js), uid)
	if err != nil {
		return fmt.Errorf("加密失败：%v", err)
	}
	err = c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

func wsMsgToJSON(wsMsg *pushgrpcv1.WsMsg, slice bool) ([]byte, error) {
	// 先将整个结构体转换为 map[string]interface{}
	msgMap := map[string]interface{}{
		"uid":          wsMsg.Uid,
		"event":        wsMsg.Event,
		"rid":          wsMsg.Rid,
		"driver_id":    wsMsg.DriverId,
		"send_at":      wsMsg.SendAt,
		"push_offline": wsMsg.PushOffline,
	}

	// 如果是切片，并且长度大于 0，则将 Data 字段解析为切片
	if slice {
		var dataSlice []map[string]interface{}
		err := json.Unmarshal(wsMsg.Data.Value, &dataSlice)
		if err != nil {
			return nil, fmt.Errorf("error decoding data field: %v", err)
		}
		msgMap["data"] = dataSlice
	} else { // 否则将 Data 字段解析为 map[string]interface{}
		var dataMap map[string]interface{}
		err := json.Unmarshal(wsMsg.Data.Value, &dataMap)
		if err != nil {
			return nil, fmt.Errorf("error decoding data field: %v", err)
		}
		msgMap["data"] = dataMap
	}

	// 将 map 转换为 JSON 格式
	jsonBytes, err := json.Marshal(msgMap)
	if err != nil {
		return nil, fmt.Errorf("error encoding to JSON: %v", err)
	}

	return jsonBytes, nil
}

// 将 PushWsBatchByUserIdsRequest 消息转换为 JSON 格式
func pushWsBatchByUserIdsRequestToJSON(req *pushgrpcv1.PushWsBatchByUserIdsRequest) ([]byte, error) {
	// 将 Data 字段的值反序列化为 map[string]interface{}
	var dataMap map[string]interface{}
	err := json.Unmarshal(req.Data.Value, &dataMap)
	if err != nil {
		return nil, fmt.Errorf("error decoding data field: %v", err)
	}

	// 构造一个 map 用于序列化为 JSON
	jsonMap := map[string]interface{}{
		"user_ids":     req.UserIds,
		"event":        req.Event,
		"data":         dataMap,
		"push_offline": req.PushOffline,
		"driver_id":    req.DriverId,
	}

	// 将 map 转换为 JSON 格式
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("error encoding PushWsBatchByUserIdsRequest to JSON: %v", err)
	}

	return jsonBytes, nil
}

// 将 PushWsBatchRequest 消息转换为 JSON 格式
func pushWsBatchRequestToJSON(req *pushgrpcv1.PushWsBatchRequest) ([]byte, error) {
	// 构造一个包含所有消息的切片，每个消息将 Data 字段的值反序列化为 map[string]interface{}
	var msgs []map[string]interface{}
	for _, msg := range req.Msgs {
		var dataMap map[string]interface{}
		err := json.Unmarshal(msg.Data.Value, &dataMap)
		if err != nil {
			return nil, fmt.Errorf("error decoding data field: %v", err)
		}

		msgMap := map[string]interface{}{
			"uid":          msg.Uid,
			"event":        msg.Event,
			"rid":          msg.Rid,
			"driver_id":    msg.DriverId,
			"send_at":      msg.SendAt,
			"data":         dataMap,
			"push_offline": msg.PushOffline,
		}

		msgs = append(msgs, msgMap)
	}

	// 构造一个 map 用于序列化为 JSON
	jsonMap := map[string]interface{}{
		"msgs": msgs,
	}

	// 将 map 转换为 JSON 格式
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("error encoding PushWsBatchRequest to JSON: %v", err)
	}

	return jsonBytes, nil
}
