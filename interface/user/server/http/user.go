package http

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/interface/user/api/model"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"time"
)

// @Summary 用户登录
// @Description 用户登录
// @Accept  json
// @Produce  json
// @param request body model.LoginRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /user/login [post]
func login(c *gin.Context) {
	req := new(model.LoginRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	// 正则表达式匹配邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		response.SetFail(c, "邮箱格式不正确", nil)
		return
	}

	resp, err := userClient.UserLogin(context.Background(), &user.UserLoginRequest{
		Email:    req.Email,
		Password: utils.HashString(req.Password),
	})
	if err != nil {
		c.Error(err)
		return
	}

	token, err := utils.GenerateToken(resp.UserId, resp.Email)
	if err != nil {
		logger.Error("failed to generate user token", zap.Error(err))
		response.SetFail(c, "token验证失败", nil)
		return
	}
	err = redisClient.Set(context.Background(), resp.UserId, token, 60*60*24*3*time.Second).Err()
	if err != nil {
		logger.Error("failed to set user token", zap.Error(err))
		response.SetFail(c, "token验证失败", nil)
		return
	}
	c.Set("user_id", resp.UserId)

	response.SetSuccess(c, "登录成功", gin.H{"token": token, "user_info": resp})
}

// @Summary 退出登录
// @Description 退出登录
// @Accept  json
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /user/logout [post]
func logout(c *gin.Context) {

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	err = redisClient.Del(context.Background(), thisId).Err()
	if err != nil {
		logger.Error("failed to set user token", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}
	c.Set("user_id", thisId)

	response.SetSuccess(c, "退出登录成功", nil)
}

// @Summary 用户注册
// @Description 用户注册
// @Accept  json
// @Produce  json
// @param request body model.RegisterRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /user/register [post]
func register(c *gin.Context) {
	req := new(model.RegisterRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	if req.Password != req.ConfirmPass {
		response.SetFail(c, "密码和确认密码不匹配", nil)
		return
	}
	// 正则表达式匹配邮箱格式
	emailRegex := regexp2.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, 0)
	if isMatch, _ := emailRegex.MatchString(req.Email); !isMatch {
		response.SetFail(c, "邮箱格式不正确", nil)
		return
	}

	//最少包括一个数字，大小字符，最短8个字符，最长20个字符
	emailRegex = regexp2.MustCompile(`^(?=.*[0-9])(?=.*[a-zA-Z]).{6,50}$`, 0)
	if isMatch, _ := emailRegex.MatchString(req.Password); !isMatch {
		response.SetFail(c, "密码格式不正确", nil)
		return
	}
	if isMatch, _ := emailRegex.MatchString(req.ConfirmPass); !isMatch {
		response.SetFail(c, "密码格式不正确", nil)
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" {
		req.Nickname = req.Email
	}

	resp, err := userClient.UserRegister(context.Background(), &user.UserRegisterRequest{
		Email:           req.Email,
		NickName:        req.Nickname,
		Password:        utils.HashString(req.Password),
		ConfirmPassword: req.ConfirmPass,
		PublicKey:       req.PublicKey,
		Avatar:          "https://fastly.picsum.photos/id/1036/200/200.jpg?hmac=Yb5E0WTltIYlUDPDqT-d0Llaaq0mJnwiCUtxx8RrtVE",
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("user_id", resp.UserId)

	response.SetSuccess(c, "注册成功", gin.H{"user_id": resp.UserId})
}

// @Summary 查询用户信息
// @Description 查询用户信息
// @Accept  json
// @Produce  json
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @Param user_id query string true "用户id"
// @Success		200 {object} model.Response{data=model.UserInfoResponse} "Status 用户状态 (0=未知状态, 1=正常状态, 2=被禁用, 3=已删除, 4=锁定状态) RelationStatus 用户关系状态 (0=不是好友, 1=是好友, 2=黑名单)"
// @Router /user/info [get]
func getUserInfo(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		response.Fail(c, "参数错误", nil)
		return
	}

	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp := &model.UserInfoResponse{}
	r, err := userClient.UserInfo(context.Background(), &user.UserInfoRequest{
		UserId: userId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	resp = &model.UserInfoResponse{
		UserId:    r.UserId,
		Nickname:  r.NickName,
		Email:     r.Email,
		Avatar:    r.Avatar,
		Signature: r.Signature,
		Status:    model.UserStatus(r.Status),
	}

	relation, err := RelationUserClient.GetUserRelation(c, &relationgrpcv1.GetUserRelationRequest{
		UserId:   thisID,
		FriendId: userId,
	})
	if err != nil {
		fmt.Printf("错误类型：%T\n", err)
		logger.Error("获取用户关系失败", zap.Error(err))
	} else {
		if relation.Status == relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED {
			resp.RelationStatus = model.UserRelationStatusFriend
		} else if relation.Status == relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
			resp.RelationStatus = model.UserRelationStatusBlacked
		} else {
			resp.RelationStatus = model.UserRelationStatusUnknown
		}
	}

	response.SetSuccess(c, "查询成功", resp)
}

// @Summary 设置用户公钥
// @Description 设置用户公钥
// @Accept json
// @Produce json
// @param request body model.SetPublicKeyRequest true "request"
// @Security BearerToken
// @Success 200 {object} model.Response{}
// @Router /user/key/set [post]
func setUserPublicKey(c *gin.Context) {
	req := new(model.SetPublicKeyRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	// 获取用户ID，可以从请求中的token中解析出来，前提是你的登录接口已经设置了正确的token
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	// 调用服务端设置用户公钥的方法
	_, err = userClient.SetUserPublicKey(context.Background(), &user.SetPublicKeyRequest{
		UserId:    thisId,
		PublicKey: req.PublicKey,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置用户公钥成功", gin.H{"public_key": ThisKey})
}

// @Summary 获取系统pgp公钥
// @Description 获取系统pgp公钥
// @Accept  json
// @Produce  json
// @Param type query model.GetType true "指定根据id还是邮箱类型查找"
// @Param email query string false "邮箱"
// @Success		200 {object} model.Response{}
// @Router /user/system/key/get [get]
func GetSystemPublicKey(c *gin.Context) {
	response.SetSuccess(c, "获取系统pgp公钥成功", gin.H{"public_key": ThisKey})
}

// @Summary 修改用户信息
// @Description 修改用户信息
// @Accept json
// @Produce json
// @param request body model.UserInfoRequest true "request"
// @Security BearerToken
// @Success 200 {object} model.Response{}
// @Router /user/info/modify [post]
func modifyUserInfo(c *gin.Context) {
	req := new(model.UserInfoRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	// 获取用户ID，可以从请求中的token中解析出来，前提是你的登录接口已经设置了正确的token
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	// 调用服务端设置用户公钥的方法
	_, err = userClient.ModifyUserInfo(context.Background(), &user.User{
		UserId:    thisId,
		NickName:  req.NickName,
		Tel:       req.Tel,
		Avatar:    req.Avatar,
		Signature: req.Signature,
		//Action:    user.UserStatus(req.Action),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "修改成功", gin.H{"user_id": thisId})
}

// @Summary 修改用户密码
// @Description 修改用户密码
// @Accept json
// @Produce json
// @param request body model.PasswordRequest true "request"
// @Security BearerToken
// @Success 200 {object} model.Response{}
// @Router /user/password/modify [post]
func modifyUserPassword(c *gin.Context) {
	req := new(model.PasswordRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	// 获取用户ID，可以从请求中的token中解析出来，前提是你的登录接口已经设置了正确的token
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	req.Password = strings.TrimSpace(req.Password)
	req.OldPasswprd = strings.TrimSpace(req.OldPasswprd)
	req.ConfirmPass = strings.TrimSpace(req.ConfirmPass)
	if req.OldPasswprd == "" {
		response.SetFail(c, "旧密码不能为空", nil)
		return
	}
	if req.Password == "" || req.ConfirmPass == "" {
		response.SetFail(c, "密码不能为空", nil)
		return
	}
	if req.Password != req.ConfirmPass {
		response.SetFail(c, "密码和确认密码不匹配", nil)
		return
	}

	pwdRegex := regexp2.MustCompile(`^(?=.*[0-9])(?=.*[a-zA-Z]).{6,50}$`, 0)
	if isMatch, _ := pwdRegex.MatchString(req.Password); !isMatch {
		response.SetFail(c, "密码格式不正确", nil)
		return
	}
	if isMatch, _ := pwdRegex.MatchString(req.ConfirmPass); !isMatch {
		response.SetFail(c, "密码格式不正确", nil)
		return
	}
	//查询用户旧密码
	info, err := userClient.GetUserPasswordByUserId(context.Background(), &user.UserRequest{
		UserId: thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	if info.Password != utils.HashString(req.OldPasswprd) {
		response.SetFail(c, "旧密码不正确", nil)
		return
	}
	// 调用服务端设置用户公钥的方法
	_, err = userClient.ModifyUserPassword(context.Background(), &user.ModifyUserPasswordRequest{
		UserId:   thisId,
		Password: utils.HashString(req.Password),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "修改成功", gin.H{"user_id": thisId})
}

// @Summary 修改用户密钥包
// @Description 修改用户密码
// @Accept json
// @Produce json
// @param request body model.ModifyUserSecretBundleRequest true "request"
// @Security BearerToken
// @Success 200 {object} model.Response{}
// @Router /user/bundle/modify [post]
func modifyUserSecretBundle(c *gin.Context) {
	req := new(model.ModifyUserSecretBundleRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	// 获取用户ID，可以从请求中的token中解析出来，前提是你的登录接口已经设置了正确的token
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = userClient.SetUserSecretBundle(context.Background(), &user.SetUserSecretBundleRequest{
		UserId:       thisId,
		SecretBundle: req.SecretBundle,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "修改成功", gin.H{"user_id": thisId})
}

// @Summary 获取用户密钥包
// @Description 获取用户密钥包
// @Accept  json
// @Produce  json
// @Param user_id query string true "用户id"
// @Success		200 {object} model.Response{}
// @Router /user/bundle/get [get]
func getUserSecretBundle(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		response.Fail(c, "参数错误", nil)
		return
	}
	_, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = userClient.UserInfo(context.Background(), &user.UserInfoRequest{
		UserId: userId,
	})
	if err != nil {
		logger.Error("获取用户信息失败", zap.Error(err))
		response.SetFail(c, "用户不存在", nil)
		return
	}

	bundle, err := userClient.GetUserSecretBundle(context.Background(), &user.GetUserSecretBundleRequest{
		UserId: userId,
	})
	if err != nil {
		return
	}
	response.SetSuccess(c, "获取成功", bundle)
}
