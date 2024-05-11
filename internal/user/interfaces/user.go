package interfaces

import (
	"fmt"
	v1 "github.com/cossim/coss-server/internal/user/api/http/v1"
	"github.com/cossim/coss-server/internal/user/app/command"
	"github.com/cossim/coss-server/internal/user/app/query"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *HttpServer) GetUser(c *gin.Context, id string) {
	getUser, err := h.app.Queries.GetUser.Handle(c, &query.GetUse{
		CurrentUser: c.Value(constants.UserID).(string),
		TargetUser:  id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取用户信息成功", getUserToResponse(getUser))
}

func getUserToResponse(e *entity.UserInfo) *v1.UserInfo {
	return &v1.UserInfo{
		Avatar:         e.Avatar,
		CossId:         e.CossID,
		Email:          e.Email,
		LastLoginTime:  e.LastLoginTime,
		LoginNumber:    int64(e.LoginNumber),
		NewDeviceLogin: e.NewDeviceLogin,
		Nickname:       e.Nickname,
		RelationStatus: e.RelationStatus.Int(),
		Signature:      e.Signature,
		Status:         v1.UserInfoStatus(e.Status),
		Tel:            e.Tel,
		UserId:         e.UserID,
		Preferences: &v1.Preferences{
			OpenBurnAfterReading:        nil,
			OpenBurnAfterReadingTimeOut: nil,
			Remark:                      nil,
			SilentNotification:          nil,
		},
	}
}

func (h *HttpServer) UpdateUser(c *gin.Context, id string) {
	var req v1.UpdateUserJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("update user password", zap.Error(err))
		return
	}

}

func (h *HttpServer) UpdateUserPassword(c *gin.Context) {
	req := &v1.UpdateUserPasswordJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	_, err := h.app.Commands.UpdatePassword.Handle(c, &command.UpdatePassword{
		UserID:          c.Value(constants.UserID).(string),
		OldPassword:     req.OldPassword,
		ConfirmPassword: req.ConfirmPassword,
		NewPassword:     req.Password,
	})
	if err != nil {
		c.Error(err)
		return
	}

}

func (h *HttpServer) UserActivate(c *gin.Context, params v1.UserActivateParams) {
	_, err := h.app.Commands.UserActivate.Handle(c, &command.UserActivate{
		UserID:           params.UserId,
		VerificationCode: params.Key,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "用户激活成功", nil)
}

func (h *HttpServer) GetUserBundle(c *gin.Context, id string) {
	getUserBundle, err := h.app.Queries.GetUserBundle.Handle(c, &query.GetUserBundle{
		UserID: id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取用户信息成功", &v1.UserSecretBundle{
		UserId:       getUserBundle.UserID,
		SecretBundle: getUserBundle.SecretBundle,
	})
}

func (h *HttpServer) UpdateUserBundle(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h *HttpServer) GetUserClients(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h *HttpServer) UserEmailVerification(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h *HttpServer) UserLogin(c *gin.Context) {
	req := &v1.UserLoginJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userLogin, err := h.app.Commands.UserLogin.Handle(c, &command.UserLogin{
		Email:       req.Email,
		Password:    req.Password,
		DriverID:    req.DriverId,
		ClientIP:    c.ClientIP(),
		DriverToken: req.DriverToken,
		Platform:    req.Platform,
	})
	if err != nil {
		fmt.Println("sfklbnsfnbjnsfjbnjsf ======> ", err)
		c.Error(err)
		return
	}

	c.Set("user_id", userLogin.UserID)
	response.SetSuccess(c, "登录成功", gin.H{"token": userLogin.Token, "user_info": ConversionUserLogin(userLogin)})
}

func ConversionUserLogin(userLogin *command.UserLoginResponse) *v1.UserInfo {
	return &v1.UserInfo{
		UserId:        userLogin.UserID,
		Nickname:      userLogin.Nickname,
		Avatar:        userLogin.Avatar,
		Signature:     userLogin.Signature,
		Status:        v1.UserInfoStatus(uint(userLogin.Status)),
		CossId:        userLogin.CossID,
		Email:         userLogin.Email,
		Tel:           userLogin.Tel,
		LastLoginTime: userLogin.LastLoginTime,
	}
}

func (h *HttpServer) UserLogout(c *gin.Context) {
	req := &v1.UserLogoutJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid := c.Value(constants.UserID).(string)
	err := h.app.Commands.UserLogout.Handle(c, &command.UserLogout{
		UserID:   uid,
		DriverID: req.DriverId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "退出登录成功", nil)
}

func (h *HttpServer) SetUserPublicKey(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h *HttpServer) ResetUserPublicKey(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h *HttpServer) UserRegister(c *gin.Context) {
	req := &v1.UserRegisterJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := h.app.Commands.UserRegister.Handle(c, &command.UserRegister{
		Email:       string(req.Email),
		Password:    req.Password,
		ConfirmPass: req.ConfirmPassword,
		Nickname:    req.Nickname,
		PublicKey:   req.PublicKey,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.Set("user_id", userID)
	response.SetSuccess(c, "注册成功", gin.H{"user_id": userID})
}

func (h *HttpServer) GetPGPPublicKey(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
