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
	_, c2, err := utils.ParseToken(token)
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
	_, c2, err := utils.ParseToken(token)
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

// @Summary websocket请求
// @Tags Push
// @Description websocket请求
// @Router /push/ws [get]
//func (h *Handler) ws(c *gin.Context) {
//	var uid string
//	var driverId string
//	token := c.Query("token")
//	//判断设备类型
//	deviceType := c.Request.Header.Get("X-Device-Type")
//	deviceType = string(constants.DetermineClientType(constants.DriverType(deviceType)))
//
//	if token == "" {
//		//id, err := pkghttp.ParseTokenReUid(c)
//		//if err != nil {
//		//	h.logger.Error("token解析失败", zap.Error(err))
//		//	return
//		//}
//		//uid = id
//		return
//	} else {
//		_, c2, err := utils.ParseToken(token)
//		if err != nil {
//			h.logger.Error("token解析失败", zap.Error(err))
//			return
//		}
//		uid = c2.UserId
//		driverId = c2.DriverId
//	}
//	if uid == "" {
//		return
//	}
//
//	//升级http请求为websocket
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		c.Error(err)
//		return
//	}
//	h.PushService.Ws(c, conn, uid, driverId, deviceType, token)
//
//}
