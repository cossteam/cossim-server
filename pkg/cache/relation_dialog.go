package cache

import (
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/redis/go-redis/v9"
)

type RelationDialogCache interface {
	GetUserDialogs(userID string) ([]*v1.Dialog, error)
	GetDialogsByIDs(dialogIDs []string) ([]*v1.Dialog, error)
	GetDialogUsersByDialogID(dialogID string) ([]string, error)
}

var _ RelationDialogCache = &RelationDialogCacheRedis{}

type RelationDialogCacheRedis struct {
	client *redis.Client
}

func (r *RelationDialogCacheRedis) GetUserDialogs(userID string) ([]*v1.Dialog, error) {
	//dialogsJSON, err := r.client.Get(userID).Result()
	//if err == redis.Nil {
	//	return nil, nil // 用户对话列表不存在，返回空切片和 nil 错误
	//} else if err != nil {
	//	return nil, err
	//}

	var dialogs []*v1.Dialog
	//err = json.Unmarshal([]byte(dialogsJSON), &dialogs)
	//if err != nil {
	//	return nil, err
	//}

	return dialogs, nil
}

func (r *RelationDialogCacheRedis) GetDialogsByIDs(dialogIDs []string) ([]*v1.Dialog, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RelationDialogCacheRedis) GetDialogUsersByDialogID(dialogID string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}
