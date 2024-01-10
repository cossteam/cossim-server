package http

import (
	"context"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
)

type LoginRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

// @Summary 用户登录
// @Description 用户登录
// @Accept  json
// @Produce  json
// @param request body LoginRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /user/login [post]
func login(c *gin.Context) {
	req := new(LoginRequest)
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
		Password: req.Password,
	})
	if err != nil {
		c.Error(err)
		return
	}
	_, err = userClient.SetUserPublicKey(context.Background(), &user.SetPublicKeyRequest{
		PublicKey: req.PublicKey,
		UserId:    resp.UserId,
	})
	if err != nil {
		return
	}

	token, err := utils.GenerateToken(resp.UserId, resp.Email)
	if err != nil {
		logger.Error("failed to generate user token", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}
	c.Set("user_id", resp.UserId)

	response.SetSuccess(c, "登录成功", gin.H{"token": token, "user_info": resp})
}

type RegisterRequest struct {
	//Email    string `json:"email" binding:"required,email"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	//ConfirmPass string `json:"confirm_password" binding:"required,eqfield=Password"`
	ConfirmPass string `json:"confirm_password" binding:"required"`
	Avatar      string `json:"avatar"`
	Nickname    string `json:"nickname"`
	PublicKey   string `json:"public_key" binding:"required"`
}

// @Summary 用户注册
// @Description 用户注册
// @Accept  json
// @Produce  json
// @param request body RegisterRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /user/register [post]
func register(c *gin.Context) {
	req := new(RegisterRequest)
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

	req.Nickname = strings.TrimSpace(req.Email)
	if req.Nickname == "" {
		req.Nickname = req.Email
	}

	resp, err := userClient.UserRegister(context.Background(), &user.UserRegisterRequest{
		Email:           req.Email,
		NickName:        req.Nickname,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPass,
		Avatar:          req.Avatar,
	})
	if err != nil {
		c.Error(err)
		return
	}
	_, err = userClient.SetUserPublicKey(context.Background(), &user.SetPublicKeyRequest{
		PublicKey: req.PublicKey,
		UserId:    resp.UserId,
	})
	if err != nil {
		return
	}
	c.Set("user_id", resp.UserId)

	response.SetSuccess(c, "注册成功", gin.H{"user_id": resp.UserId})
}

type GetType int

const (
	EmailType GetType = iota
	UserIdType
)

// @Summary 查询用户信息
// @Description 查询用户信息
// @Accept  json
// @Produce  json
// @Param user_id query string true "用户id"
// @Param type query GetType true "指定根据id还是邮箱类型查找"
// @Param email query string false "邮箱"
// @Success		200 {object} utils.Response{}
// @Router /user/info [get]
func GetUserInfo(c *gin.Context) {
	email := c.Query("email")
	getType := c.Query("type")
	userId := c.Query("user_id")
	if email == "" && userId == "" {
		response.Fail(c, "参数错误", nil)
		return
	}
	gtype, _ := strconv.Atoi(getType)

	switch GetType(gtype) {
	case EmailType:
		// 正则表达式匹配邮箱格式
		emailRegex := regexp2.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, 0)
		if isMatch, _ := emailRegex.MatchString(email); !isMatch {
			response.SetFail(c, "邮箱格式不正确", nil)
			return
		}
		resp, err := userClient.GetUserInfoByEmail(context.Background(), &user.GetUserInfoByEmailRequest{
			Email: email,
		})
		if err != nil {
			c.Error(err)
			return
		}
		if resp == nil {
			response.SetFail(c, "用户不存在", nil)
			return
		}
		response.SetSuccess(c, "查询成功", resp)
	case UserIdType:
		resp, err := userClient.UserInfo(context.Background(), &user.UserInfoRequest{
			UserId: userId,
		})
		if err != nil {
			c.Error(err)
		}
		if resp == nil {
			response.SetFail(c, "用户不存在", nil)
			return
		}
		response.SetSuccess(c, "查询成功", resp)
	default:
		response.SetFail(c, "参数错误", nil)
	}
}

type SetPublicKeyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
}

// @Summary 设置用户公钥
// @Description 设置用户公钥
// @Accept json
// @Produce json
// @param request body SetPublicKeyRequest true "request"
// @Security BearerToken
// @Success 200 {object} utils.Response{}
// @Router /user/key/set [post]
func setUserPublicKey(c *gin.Context) {
	req := new(SetPublicKeyRequest)
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
// @Param type query GetType true "指定根据id还是邮箱类型查找"
// @Param email query string false "邮箱"
// @Success		200 {object} utils.Response{}
// @Router /user/system/key/get [get]
func GetSystemPublicKey(c *gin.Context) {
	response.SetSuccess(c, "获取系统pgp公钥成功", gin.H{"public_key": ThisKey})
}
