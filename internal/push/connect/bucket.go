package connect

import (
	"sync"
)

type Bucket struct {
	cLock  sync.RWMutex // protect the channels for chs
	indexs []*UserIndex
}

// 构造方法
func NewBucket() *Bucket {
	return &Bucket{
		cLock:  sync.RWMutex{},
		indexs: make([]*UserIndex, 0),
	}
}

// push 方法用于将长连接添加到 Bucket 中
func (b *Bucket) Push(client *UserIndex) {
	b.cLock.Lock()
	defer b.cLock.Unlock()

	b.indexs = append(b.indexs, client)
}

// get 方法用于获取 Bucket 中的用户索引列表
func (b *Bucket) Get() []*UserIndex {
	b.cLock.RLock()
	defer b.cLock.RUnlock()

	return b.indexs
}

func (b *Bucket) GetUserIndex(userID string) *UserIndex {
	for _, index := range b.indexs {
		if index.UserId == userID {
			return index
		}
	}
	return nil
}

func (b *Bucket) GetLength() int {
	return len(b.indexs)
}

// DeleteByUserID 方法用于根据 UserID 从 Bucket 中删除对应的 UserIndex 对象
func (b *Bucket) DeleteByUserID(userID string) {
	b.cLock.Lock()
	defer b.cLock.Unlock()

	for i, idx := range b.indexs {
		if idx.UserId == userID {
			// 通过将切片中对应元素与最后一个元素交换位置，然后缩减切片长度的方式删除元素
			b.indexs[i] = b.indexs[len(b.indexs)-1]
			b.indexs = b.indexs[:len(b.indexs)-1]
			return
		}
	}
}

func (b *Bucket) SendMessage(userId string, message string) error {
	if b.GetUserIndex(userId) == nil {
		return nil
	}
	err := b.GetUserIndex(userId).SendMessage(message)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bucket) DelUserWsClient(userId string, rid int64) {
	b.GetUserIndex(userId).DelUserWsClient(rid)
}
