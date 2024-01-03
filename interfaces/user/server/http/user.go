package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	user "im/services/user/api/v1"
	"net/http"
	"regexp"
)

func login(c *gin.Context) {
	type LoginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	req := new(LoginRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "参数验证失败"})
		return
	}

	resp, err := userClient.UserLogin(context.Background(), &user.UserLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	//todo 生成token
	if err != nil {
		logger.Error("user service failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "登录成功", "data": resp})
}

func register(c *gin.Context) {
	type RegisterRequest struct {
		//Email    string `json:"email" binding:"required,email"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		//ConfirmPass string `json:"confirm_password" binding:"required,eqfield=Password"`
		ConfirmPass string `json:"confirm_password" binding:"required"`
		Avatar      string `json:"avatar"`
		Nickname    string `json:"nickname"`
	}
	req := new(RegisterRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "参数验证失败"})
		return
	}

	if req.Password != req.ConfirmPass {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "密码和确认密码不匹配"})
		return
	}

	// 正则表达式匹配邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "邮箱格式不正确"})
		return
	}
	//最少包括一个数字，大小字符，最短8个字符，最长20个字符
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9$@$!%*?&]{8,20}$`)
	if !emailRegex.MatchString(req.Password) || !emailRegex.MatchString(req.ConfirmPass) {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "密码格式不正确"})
		return
	}

	resp, err := userClient.UserRegister(context.Background(), &user.UserRegisterRequest{
		Email:           req.Email,
		NickName:        req.Nickname,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPass,
		Avatar:          req.Avatar,
	})
	if err != nil {
		logger.Error("user service failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "注册成功", "data": map[string]interface{}{
		"userID": resp.UserId,
	}})
}
