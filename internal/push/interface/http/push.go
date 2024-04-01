package http

import (
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary websocket请求
// @Tags Msg
// @Description websocket请求
// @Router /msg/ws [get]
func (h *Handler) ws(c *gin.Context) {
	var uid string
	var driverId string
	token := c.Query("token")
	//判断设备类型
	deviceType := c.Request.Header.Get("X-Device-Type")
	deviceType = string(constants.DetermineClientType(constants.DriverType(deviceType)))

	if token == "" {
		//id, err := pkghttp.ParseTokenReUid(c)
		//if err != nil {
		//	h.logger.Error("token解析失败", zap.Error(err))
		//	return
		//}
		//uid = id
		return
	} else {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			h.logger.Error("token解析失败", zap.Error(err))
			return
		}
		uid = c2.UserId
		driverId = c2.DriverId
	}
	if uid == "" {
		return
	}

	//升级http请求为websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(err)
		return
	}
	h.PushService.Ws(conn, uid, driverId, deviceType, token)

}
