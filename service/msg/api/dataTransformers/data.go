package dataTransformers

import "github.com/cossim/coss-server/service/msg/domain/entity"

type UserMsgListResponse struct {
	UserMessages []entity.UserMessage `json:"user_messages" form:"user_messages" uri:"user_messages"`
	Total        int32                `json:"total" form:"total" uri:"total"`
	CurrentPage  int32                `json:"current_page" form:"current_page" uri:"current_page"`
}
