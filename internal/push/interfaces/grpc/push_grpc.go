package grpc

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/utils"
)

func (h *Handler) Push(ctx context.Context, request *v1.PushRequest) (*v1.PushResponse, error) {
	resp := &v1.PushResponse{}
	switch request.Type {
	case v1.Type_Ws:
		bytes := request.GetData()
		wsMsg2 := &v1.WsMsg{}
		if err := utils.BytesToStruct(bytes, wsMsg2); err != nil {
			fmt.Println("Error while deserializing WsMsg:", err)
			return resp, err
		}

		_, err := h.PushService.PushWs(ctx, wsMsg2)
		if err != nil {
			return nil, err
		}
	case v1.Type_Ws_Batch:
		bytes := request.GetData()
		wsMsg2 := &v1.PushWsBatchRequest{}
		if err := utils.BytesToStruct(bytes, wsMsg2); err != nil {
			fmt.Println("Error while deserializing WsMsg:", err)
			return resp, err
		}

		_, err := h.PushService.PushWsBatch(ctx, wsMsg2)
		if err != nil {
			return nil, err
		}
	case v1.Type_Ws_Batch_User:
		bytes := request.GetData()
		wsMsg2 := &v1.PushWsBatchByUserIdsRequest{}
		if err := utils.BytesToStruct(bytes, wsMsg2); err != nil {
			fmt.Println("Error while deserializing WsMsg:", err)
			return resp, err
		}

		_, err := h.PushService.PushWsBatchByUserIds(ctx, wsMsg2)
		if err != nil {
			return nil, err
		}
	case v1.Type_Mobile:
		//移动推送
	case v1.Type_Email:
		//发送邮件
	case v1.Type_Message:
		//发送短信
	}

	return resp, nil
}
