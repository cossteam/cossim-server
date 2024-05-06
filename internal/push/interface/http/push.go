package http

import (
	"context"
	"github.com/cossim/coss-server/pkg/utils"
	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
)

// @Summary websocket请求
// @Tags Push
// @Description websocket请求
// @Router /push/ws [get]
func (h *Handler) ws(s socketio.Conn) error {
	url := s.URL()
	token := url.Query().Get("token")
	_, c2, err := utils.ParseToken(token, h.jwtSecret)
	if err != nil {
		return err
	}
	userId := c2.UserId

	s.Join(userId)
	s.SetContext(nil)
	err = h.PushService.Ws(context.Background(), s, userId, c2.DriverId, s.ID(), token)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) disconnect(s socketio.Conn, msg string) {
	url := s.URL()
	token := url.Query().Get("token")
	_, c2, err := utils.ParseToken(token, h.jwtSecret)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		return
	}
	userId := c2.UserId
	err = h.PushService.WsOfflineClients(context.Background(), userId, s.ID())
	if err != nil {
		h.logger.Error("推送离线消息失败", zap.Error(err))
	}

}

func (h *Handler) reply(s socketio.Conn, msg string) {
	s.Emit("reply", "服务端触发客户端事件： "+msg)
}

func (h *Handler) bye(s socketio.Conn) string {
	last := s.Context().(string)
	s.Emit("bye", last)
	s.Close()
	return last
}

func (h *Handler) error(s socketio.Conn, e error) {
	h.logger.Error("socketio error", zap.Error(e))
}
