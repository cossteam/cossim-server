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

func (h *HttpServer) SearchUser(c *gin.Context, params v1.SearchUserParams) {
	h.logger.Info("Search user", zap.String("email", params.Email))
	searchUser, err := h.app.Queries.GetUser.Handle(c, &query.GetUse{
		CurrentUser: c.Value(constants.UserID).(string),
		//TargetUser:  params.Email,
		TargetEmail: params.Email,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "User found", getUserToResponse(searchUser))
}

func (h *HttpServer) UpdateUserAvatar(c *gin.Context) {
	// Parse form data
	if err := c.Request.ParseMultipartForm(25 << 20); // 25 MB limit
	err != nil {
		response.SetFail(c, "Failed to parse form data", nil)
		return
	}

	// Get the file from the form data
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		response.SetFail(c, "Error retrieving the file", nil)
		return
	}
	defer file.Close()

	// Check file type
	contentType := handler.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		response.SetFail(c, "Unsupported file type. Only JPEG and PNG are allowed.", nil)
		return
	}

	// Check file size
	if handler.Size > 25<<20 { // 25 MB limit
		response.SetFail(c, "File size exceeds the limit. Maximum allowed size is 25 MB.", nil)
		return
	}

	url, err := h.app.Commands.UpdateUserAvatarHandler.Handle(c, &command.UpdateUserAvatar{
		UserID: c.Value(constants.UserID).(string),
		Avatar: file,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新用户头像成功", gin.H{"avatar": url})
}

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
	var preferences *v1.Preferences
	if e.Preferences != nil {
		preferences = &v1.Preferences{
			OpenBurnAfterReading:        e.Preferences.OpenBurnAfterReading,
			OpenBurnAfterReadingTimeOut: int(e.Preferences.OpenBurnAfterReadingTimeOut),
			Remark:                      e.Preferences.Remark,
			SilentNotification:          e.Preferences.SilentNotification,
		}
	}
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
		Preferences:    preferences,
	}
}

func (h *HttpServer) UpdateUser(c *gin.Context) {
	var req v1.UpdateUserJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("update user password", zap.Error(err))
		return
	}

	_, err := h.app.Commands.UpdateUser.Handle(c, &command.UpdateUser{
		UserID:    c.Value(constants.UserID).(string),
		Avatar:    req.Avatar,
		CossID:    req.CossId,
		Nickname:  req.Nickname,
		Signature: req.Signature,
		Tel:       req.Tel,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新用户信息成功", nil)
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

	response.SetSuccess(c, "更新用户密码成功", nil)
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

	response.SetSuccess(c, "获取用户密钥包成功", &v1.UserSecretBundle{
		UserId:       getUserBundle.UserID,
		SecretBundle: getUserBundle.SecretBundle,
	})
}

func (h *HttpServer) UpdateUserBundle(c *gin.Context) {
	req := &v1.UpdateUserBundleJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	err := h.app.Commands.UpdateUserBundle.Handle(c, &command.UpdateUserBundle{
		UserID:       c.Value(constants.UserID).(string),
		SecretBundle: req.SecretBundle,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新用户密钥包成功", nil)
}

func (h *HttpServer) GetUserLoginClients(c *gin.Context) {
	getUserClients, err := h.app.Queries.GetUserLoginClients.Handle(c, &query.GetUserLoginClients{
		UserID: c.Value(constants.UserID).(string),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取用户客户端成功", getUserClientsToResponse(getUserClients))
}

func getUserClientsToResponse(e []*query.GetUserLoginClientsResponse) []*v1.UserLoginClient {
	var userClients []*v1.UserLoginClient
	for _, v := range e {
		userClients = append(userClients, &v1.UserLoginClient{
			ClientIp:   v.ClientIP,
			DriverId:   v.DriverID,
			DriverType: v.DriverType,
			LoginAt:    v.LoginAt,
		})
	}
	return userClients
}

func (h *HttpServer) UserEmailVerification(c *gin.Context) {
	req := &v1.UserEmailVerificationJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	err := h.app.Commands.SendUserEmailVerification.Handle(c, &command.SendUserEmailVerification{
		//UserID: c.Value(constants.UserID).(string),
		Email: string(req.Email),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送邮箱验证码成功", nil)
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
		c.Error(err)
		return
	}

	userInfo, err := h.app.Queries.GetUser.Handle(c, &query.GetUse{
		TargetUser: userLogin.UserID,
	})
	if err != nil {
		fmt.Println("userlogin ======> ", err)
		c.Error(err)
		return
	}

	c.Set(constants.UserID, userLogin.UserID)
	c.Set(constants.PublicKey, userInfo.PublicKey)
	response.SetSuccess(c, "登录成功", gin.H{"token": userLogin.Token, "user_info": ConversionUserLogin(userLogin)})
}

func ConversionUserLogin(userLogin *command.UserLoginResponse) *v1.UserInfo {
	return &v1.UserInfo{
		UserId:         userLogin.UserID,
		Nickname:       userLogin.Nickname,
		Avatar:         userLogin.Avatar,
		Signature:      userLogin.Signature,
		CossId:         userLogin.CossID,
		Email:          userLogin.Email,
		Tel:            userLogin.Tel,
		Status:         v1.UserInfoStatus(uint(userLogin.Status)),
		NewDeviceLogin: userLogin.NewDeviceLogin,
		LastLoginTime:  userLogin.LastLoginTime,
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
	req := &v1.SetUserPublicKeyJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	err := h.app.Commands.SetUserPublicKey.Handle(c, &command.SetUserPublicKey{
		UserID:    c.Value(constants.UserID).(string),
		PublicKey: req.PublicKey,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置用户公钥成功", nil)
}

func (h *HttpServer) ResetUserPublicKey(c *gin.Context) {
	req := &v1.ResetUserPublicKeyJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	err := h.app.Commands.ResetUserPublicKey.Handle(c, &command.ResetUserPublicKey{
		UserID:    c.Value(constants.UserID).(string),
		PublicKey: req.PublicKey,
		Code:      req.Code,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "重置用户公钥成功", nil)
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

	userInfo, err := h.app.Queries.GetUser.Handle(c, &query.GetUse{
		TargetUser: userID,
	})
	if err != nil {
		fmt.Println("userlogin ======> ", err)
		c.Error(err)
		return
	}

	c.Set(constants.UserID, userID)
	c.Set(constants.PublicKey, userInfo.PublicKey)
	response.SetSuccess(c, "注册成功", gin.H{"user_id": userID})
}

func (h *HttpServer) GetPGPPublicKey(c *gin.Context) {
	response.SetSuccess(c, "获取系统pgp公钥成功", gin.H{"public_key": h.pgpKey})
}
