package service

import (
	"context"
	"encoding/json"
	"fmt"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/api/http/model"
	"github.com/cossim/coss-server/pkg/utils/time"
	"go.uber.org/zap"
	ostime "time"
)

func (s *Service) OpenOrCloseDialog(ctx context.Context, userId string, request *model.CloseOrOpenDialogRequest) (interface{}, error) {
	_, err := s.svc.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.svc.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: request.DialogId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.svc.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
		DialogId: request.DialogId,
		UserId:   userId,
		Action:   relationgrpcv1.CloseOrOpenDialogType(request.Action),
	})
	if err != nil {
		return nil, err
	}
	//if request.Action == model.CloseDialog {
	//	err := s.removeRedisUserDialogList(userId, request.DialogId)
	//	if err != nil {
	//		return nil, err
	//	}
	//} else {
	//	id, err := s.svc.GetDialogTargetUserId(ctx, &relationgrpcv1.GetDialogTargetUserIdRequest{
	//		DialogId: request.DialogId,
	//		UserId:   userId,
	//	})
	//	if err != nil {
	//		return nil, err
	//	}
	//	if len(id.UserIds) == 1 {
	//		info2, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
	//			UserId: id.UserIds[0],
	//		})
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
	//			UserId: userId,
	//		})
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
	//			UserId:   userId,
	//			FriendId: info2.UserId,
	//		})
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		name := info2.NickName
	//		if relation.Remark != "" {
	//			name = relation.Remark
	//		}
	//
	//		//查询最后一条消息
	//		lastMsg, err := s.msgClient.GetLastMsgsByDialogIds(ctx, &msggrpcv1.GetLastMsgsByDialogIdsRequest{
	//			DialogIds: []uint32{request.DialogId},
	//		})
	//
	//		lm := lastMsg.LastMsgs[0]
	//
	//		lastmsg := model.Message{
	//			MsgType:            uint(lm.Type),
	//			Content:            lm.Content,
	//			SenderId:           lm.SenderId,
	//			SendTime:           lm.CreatedAt,
	//			MsgId:              uint64(lm.Id),
	//			IsBurnAfterReading: model.BurnAfterReadingType(lm.IsBurnAfterReadingType),
	//			IsLabel:            model.LabelMsgType(lm.IsLabel),
	//			ReplyId:           lm.ReplyId,
	//		}
	//		if lm.SenderId == info.UserId {
	//			lastmsg.SenderInfo = model.SenderInfo{
	//				Name:   info.NickName,
	//				Avatar: info.Avatar,
	//				UserId: info.UserId,
	//			}
	//			lastmsg.ReceiverInfo = model.SenderInfo{
	//				Name:   info2.NickName,
	//				Avatar: info2.Avatar,
	//				UserId: info2.UserId,
	//			}
	//		} else {
	//			lastmsg.SenderInfo = model.SenderInfo{
	//				Name:   info2.NickName,
	//				Avatar: info2.Avatar,
	//				UserId: info2.UserId,
	//			}
	//			lastmsg.ReceiverInfo = model.SenderInfo{
	//				Name:   info.NickName,
	//				Avatar: info.Avatar,
	//				UserId: info.UserId,
	//			}
	//		}
	//
	//		re := model.UserDialogListResponse{
	//			DialogId:       dialogInfo.Id,
	//			UserId:         info2.UserId,
	//			DialogType:     model.ConversationType(dialogInfo.Type),
	//			DialogName:     name,
	//			DialogAvatar:   info2.Avatar,
	//			DialogCreateAt: dialogInfo.CreateAt,
	//			TopAt:          int64(dialogUser.TopAt),
	//			LastMessage:    lastmsg,
	//		}
	//
	//		err = s.insertRedisUserDialogList(userId, re)
	//		if err != nil {
	//			return nil, err
	//		}
	//	}
	//
	//}

	err = s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userId))
	if err != nil {
		s.logger.Error("删除用户好友失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) TopOrCancelTopDialog(ctx context.Context, userId string, request *model.TopOrCancelTopDialogRequest) (interface{}, error) {
	_, err := s.svc.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.svc.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: request.DialogId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.svc.TopOrCancelTopDialog(ctx, &relationgrpcv1.TopOrCancelTopDialogRequest{
		DialogId: request.DialogId,
		UserId:   userId,
		Action:   relationgrpcv1.TopOrCancelTopDialogType(request.Action),
	})
	if err != nil {
		return nil, err
	}

	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(fmt.Sprintf("dialog:%s", userId))
	if err != nil {
		return nil, err
	}

	if !f {
		return nil, nil
	}
	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("dialog:%s", userId))
	if err != nil {
		return nil, err
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []model.UserDialogListResponse
		for _, v := range values {
			// 在这里根据实际的数据结构进行解析
			// 这里假设你的缓存数据是 JSON 字符串，需要解析为 UserDialogListResponse 类型
			var dialog model.UserDialogListResponse
			err := json.Unmarshal([]byte(v), &dialog)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				continue
			}
			if dialog.DialogId == request.DialogId {
				if request.Action == model.TopDialog {
					dialog.TopAt = time.Now()
				} else {
					dialog.TopAt = 0
				}
			}
			responseList = append(responseList, dialog)
		}

		//保存回redis
		// 创建一个新的[]interface{}类型的数组
		var interfaceList []interface{}

		// 遍历responseList数组，并将每个元素转换为interface{}类型后添加到interfaceList数组中
		for _, dialog := range responseList {
			interfaceList = append(interfaceList, dialog)
		}

		err := s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userId))
		if err != nil {
			return nil, err
		}

		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("dialog:%s", userId), interfaceList)
		if err != nil {
			return nil, err
		}
		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("dialog:%s", userId), 3*24*ostime.Hour)
		if err != nil {
			return nil, err
		}

	}

	return nil, nil
}
