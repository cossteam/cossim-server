package cache

import (
	"context"
	"encoding/json"
	"fmt"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const (
	MsgExpireTime  = 12 * time.Hour
	MsgKeyPrefix   = "msg:"
	MsgLastMessage = MsgKeyPrefix + "last_msg:"
)

func GetLastMessageKey(dialogID uint32) string {
	return MsgLastMessage + strconv.FormatUint(uint64(dialogID), 10)
}

type MsgCache interface {
	GetLastMessageByDialogIDs(ctx context.Context, dialogIds ...uint32) (*msggrpcv1.GetLastMsgsResponse, error)
	SetLastMessage(ctx context.Context, dialogID uint32, lastMsg *msggrpcv1.LastMsg, expiration time.Duration) error
	DeleteLastMessage(ctx context.Context, dialogIDs ...uint32) error
}

func NewMsgCacheRedis(addr, password string, db int) (*MsgCacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &MsgCacheRedis{
		client: client,
	}, nil
}

var _ MsgCache = &MsgCacheRedis{}

type MsgCacheRedis struct {
	client *redis.Client
}

func (m *MsgCacheRedis) SetLastMessage(ctx context.Context, dialogID uint32, lastMsg *msggrpcv1.LastMsg, expiration time.Duration) error {
	if lastMsg == nil {
		return ErrCacheContentEmpty
	}

	key := GetLastMessageKey(dialogID)
	lastMsgJSON, err := json.Marshal(lastMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal last message: %v", err)
	}

	err = m.client.Set(ctx, key, lastMsgJSON, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set last message in cache: %v", err)
	}

	return nil
}

func (m *MsgCacheRedis) DeleteLastMessage(ctx context.Context, dialogIDs ...uint32) error {
	keys := make([]string, 0, len(dialogIDs))
	for _, did := range dialogIDs {
		keys = append(keys, GetLastMessageKey(did))
	}
	return m.client.Del(ctx, keys...).Err()
}

func (m *MsgCacheRedis) GetLastMessageByDialogIDs(ctx context.Context, dialogIds ...uint32) (*msggrpcv1.GetLastMsgsResponse, error) {
	if len(dialogIds) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	keys := make([]string, len(dialogIds))
	for i, did := range dialogIds {
		keys[i] = GetLastMessageKey(did)
	}

	vals, err := m.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	resp := &msggrpcv1.GetLastMsgsResponse{}
	for _, val := range vals {
		if val == nil {
			continue
		}

		relation := &msggrpcv1.LastMsg{}
		if err := json.Unmarshal([]byte(val.(string)), relation); err != nil {
			return nil, err
		}

		resp.LastMsgs = append(resp.LastMsgs, relation)
	}

	return resp, nil
}
