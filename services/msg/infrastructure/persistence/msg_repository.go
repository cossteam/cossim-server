package persistence

import (
	"github.com/cossim/coss-server/services/msg/domain/entity"
	"gorm.io/gorm"
)

type MsgRepo struct {
	db *gorm.DB
}

func NewMsgRepo(db *gorm.DB) *MsgRepo {
	return &MsgRepo{db: db}
}

func (g *MsgRepo) InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint) (*entity.UserMessage, error) {
	content := &entity.UserMessage{
		SendID:    senderId,
		ReceiveID: receiverId,
		Content:   msg,
		Type:      msgType,
		ReplyId:   replyId,
	}
	if err := g.db.Save(content).Error; err != nil {
		return nil, err
	}
	return content, nil
}

func (g *MsgRepo) InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint) (*entity.GroupMessage, error) {
	content := &entity.GroupMessage{
		UID:     uid,
		GroupID: groupId,
		Content: msg,
		Type:    msgType,
		ReplyId: replyId,
	}
	if err := g.db.Save(content).Error; err != nil {
		return nil, err
	}
	return content, nil
}

func (g *MsgRepo) GetUserMsgList(userId, friendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]entity.UserMessage, int32, int32) {
	var results []entity.UserMessage

	query := g.db.Model(&entity.UserMessage{})
	query = query.Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)", userId, friendId, friendId, userId)

	if content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	if entity.IsValidMessageType(msgType) {
		query = query.Where("msg_type = ?", msgType)
	}

	offset := (pageNumber - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize).Find(&results)
	var count int64
	// 注意：这里的 query 是一个新的查询，需要重新构建条件
	countQuery := g.db.Model(&entity.UserMessage{})
	countQuery = countQuery.Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)", userId, friendId, friendId, userId)
	if content != "" {
		countQuery = countQuery.Where("content LIKE ?", "%"+content+"%")
	}
	if entity.IsValidMessageType(msgType) {
		countQuery = countQuery.Where("msg_type = ?", msgType)
	}
	countQuery.Count(&count)

	return results, int32(count), int32(pageNumber)

}
