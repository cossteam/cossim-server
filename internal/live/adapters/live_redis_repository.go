package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/live/domain/live"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	liveUserPrefix  = "live.User."
	liveGroupPrefix = "live.Group."
	liveRoomPrefix  = "live.Room."
)

var (
	timeout              = 60 * time.Second
	ErrCacheContentEmpty = errors.New("cache content cannot be empty")
	ErrCacheKeyEmpty     = errors.New("cache key cannot be empty")
)

var _ live.Repository = &RedisLiveRepository{}

func NewRedisLiveRepository(addr, password string, db int) (*RedisLiveRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisLiveRepository{
		client: client,
	}, nil
}

type RedisLiveRepository struct {
	client *redis.Client
}

func (r *RedisLiveRepository) SetRoomPersist(ctx context.Context, roomID string) error {
	return r.setPersist(ctx, liveRoomPrefix+roomID)
}

func (r *RedisLiveRepository) SetUserLivePersist(ctx context.Context, userID string) error {
	return r.setPersist(ctx, liveUserPrefix+userID)
}

func (r *RedisLiveRepository) SetGroupLivePersist(ctx context.Context, groupID string) error {
	return r.setPersist(ctx, liveGroupPrefix+groupID)
}

func (r *RedisLiveRepository) setPersist(ctx context.Context, key string) error {
	if key == "" {
		return ErrCacheKeyEmpty
	}

	remaining, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	if remaining < 0 {
		return code.LiveErrCallNotFound
	}

	// 删除过期时间，使键永久保持有效
	return r.client.Persist(ctx, key).Err()
}

func (r *RedisLiveRepository) UpdateGroupLiveExpiration(ctx context.Context, groupID string, expiration time.Duration) error {
	//if roomID == "" || groupID == "" {
	//	return ErrCacheKeyEmpty
	//}
	//
	//remaining, err := r.client.TTL(ctx, liveGroupPrefix+groupID).Result()
	//if err != nil {
	//	return err
	//}
	//
	//if remaining < 0 {
	//	return code.LiveErrCallNotFound
	//}
	//if expiration == -1 {
	//	expiration = remaining
	//}
	//
	//return r.client.Set(ctx, liveUserPrefix+groupID, roomID, expiration).Err()

	return r.setKeyExpiration(ctx, liveGroupPrefix+groupID, expiration)
}

func (r *RedisLiveRepository) setKeyExpiration(ctx context.Context, key string, expiration time.Duration) error {
	if key == "" {
		return ErrCacheKeyEmpty
	}

	remaining, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	if remaining < 0 {
		return code.LiveErrCallNotFound
	}
	if expiration == -1 {
		expiration = remaining
	}

	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisLiveRepository) UpdateUserLiveExpiration(ctx context.Context, userID string, expiration time.Duration) error {
	//// 检查房间ID和用户ID是否为空
	//if roomID == "" || userID == "" {
	//	return ErrCacheKeyEmpty
	//}
	//
	//remaining, err := r.client.TTL(ctx, liveUserPrefix+userID).Result()
	//if err != nil {
	//	return err
	//}
	//
	//if remaining < 0 {
	//	return code.LiveErrCallNotFound
	//}
	//if expiration == -1 {
	//	expiration = remaining
	//}
	//
	//return r.client.Set(ctx, liveUserPrefix+userID, roomID, expiration).Err()

	return r.setKeyExpiration(ctx, liveUserPrefix+userID, expiration)
}

func (r *RedisLiveRepository) UpdateRoomWithExpiration(ctx context.Context, room *live.Room, expiration time.Duration) error {
	remaining, err := r.client.TTL(ctx, liveRoomPrefix+room.ID).Result()
	if err != nil {
		return err
	}

	if remaining < 0 && expiration != 0 {
		return code.LiveErrCallNotFound
	}
	if expiration == -1 {
		expiration = remaining
	}

	data, err := room.Marshal()
	if err != nil {
		return err
	}

	return r.client.Set(ctx, liveRoomPrefix+room.ID, data, expiration).Err()
}

func (r *RedisLiveRepository) CreateGroupLive(ctx context.Context, roomID string, groupID string) error {
	return r.client.Set(ctx, liveGroupPrefix+groupID, roomID, timeout).Err()
}

func (r *RedisLiveRepository) DeleteGroupLive(ctx context.Context, groupID string) error {
	return r.client.Del(ctx, liveGroupPrefix+groupID).Err()
}

func (r *RedisLiveRepository) GetGroupRoom(ctx context.Context, groupID string) (*live.Room, error) {
	roomID, err := r.client.Get(ctx, liveGroupPrefix+groupID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, code.LiveErrCallNotFound.Reason(err)
		}
		return nil, err
	}

	room, err := r.client.Get(ctx, liveRoomPrefix+roomID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, code.LiveErrCallNotFound.Reason(err)
		}
		return nil, err
	}

	resp := &live.Room{}
	err = resp.Unmarshal([]byte(room))
	return resp, err
}

func (r *RedisLiveRepository) CreateUsersLive(ctx context.Context, roomID string, userIDs ...string) error {
	if roomID == "" {
		return ErrCacheKeyEmpty
	}
	if len(userIDs) == 0 {
		return ErrCacheContentEmpty
	}

	// Prepare the values slice with user IDs and room ID pairs
	values := make([]interface{}, 0, len(userIDs)*2)
	for _, userID := range userIDs {
		values = append(values, liveUserPrefix+userID, roomID)
	}

	// Use MSet method to set multiple keys and values
	if err := r.client.MSet(ctx, values...).Err(); err != nil {
		return err
	}
	// Set expiration for each key
	for i := 0; i < len(userIDs); i++ {
		err := r.client.Expire(ctx, liveUserPrefix+userIDs[i], timeout).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisLiveRepository) DeleteUsersLive(ctx context.Context, userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}
	if len(userIDs) == 0 {
		return ErrCacheContentEmpty
	}

	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = liveUserPrefix + id
	}

	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisLiveRepository) GetUserRooms(ctx context.Context, userID string) ([]*live.Room, error) {
	room, err := r.client.Get(ctx, liveUserPrefix+userID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, code.LiveErrCallNotFound.Reason(err)
		}
		return nil, fmt.Errorf("%w", err)
	}

	getRoom, err := r.GetRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	//resp := &live.Room{}
	//if err = resp.Unmarshal([]byte(room)); err != nil {
	//	return nil, err
	//}
	return []*live.Room{getRoom}, nil
}

func (r *RedisLiveRepository) CreateRoom(ctx context.Context, room *live.Room) error {
	return r.client.Set(ctx, liveRoomPrefix+room.ID, room.String(), timeout).Err()
}

func (r *RedisLiveRepository) GetRoom(ctx context.Context, id string) (*live.Room, error) {
	room, err := r.client.Get(ctx, liveRoomPrefix+id).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, code.LiveErrCallNotFound.Reason(err)
		}
		return nil, fmt.Errorf("%w", err)
	}

	resp := &live.Room{}
	if err = resp.Unmarshal([]byte(room)); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *RedisLiveRepository) UpdateRoom(ctx context.Context, room *live.Room) error {
	if room == nil {
		return ErrCacheContentEmpty
	}

	roomKey := liveRoomPrefix + room.ID

	ttl, err := r.client.TTL(ctx, roomKey).Result()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	data, err := room.Marshal()
	if err != nil {
		return err
	}

	return r.client.Set(ctx, roomKey, data, ttl).Err()
}

func (r *RedisLiveRepository) DeleteRoom(ctx context.Context, roomID string) error {
	return r.client.Del(ctx, liveRoomPrefix+roomID).Err()
}

func (r *RedisLiveRepository) ListRooms(ctx context.Context, roomIDs []string) ([]*live.Room, error) {
	if len(roomIDs) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	keys := make([]string, len(roomIDs))
	for i, id := range roomIDs {
		keys[i] = liveRoomPrefix + id
	}

	rms, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	rooms := make([]*live.Room, 0, len(roomIDs))
	for _, rm := range rms {
		if rm == nil {
			continue
		}

		room := &live.Room{}
		if err := json.Unmarshal([]byte(rm.(string)), room); err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *RedisLiveRepository) GetParticipant(ctx context.Context, roomID string, participantID string) (*live.ParticipantInfo, error) {
	room, err := r.client.Get(ctx, liveRoomPrefix+roomID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, code.LiveErrCallNotFound.Reason(err)
		}
		return nil, fmt.Errorf("%w", err)
	}

	resp := &live.Room{}
	if err = resp.Unmarshal([]byte(room)); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *RedisLiveRepository) ListRoomsByCreator(ctx context.Context, creator string) ([]*live.Room, error) {
	var rooms []*live.Room

	// Get all room keys with the given creator prefix
	creatorKey := liveRoomPrefix + "creator:" + creator
	roomKeys, err := r.client.Keys(ctx, creatorKey).Result()
	if err != nil {
		return nil, err
	}

	// Retrieve room data for each room key
	for _, key := range roomKeys {
		roomData, err := r.client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		room := &live.Room{}
		if err := json.Unmarshal([]byte(roomData), room); err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *RedisLiveRepository) ListRoomsByOwner(ctx context.Context, owner string) ([]*live.Room, error) {
	var rooms []*live.Room

	// Get all room keys with the given owner prefix
	ownerKey := liveRoomPrefix + "owner:" + owner
	roomKeys, err := r.client.Keys(ctx, ownerKey).Result()
	if err != nil {
		return nil, err
	}

	// Retrieve room data for each room key
	for _, key := range roomKeys {
		roomData, err := r.client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		room := &live.Room{}
		if err := json.Unmarshal([]byte(roomData), room); err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}
