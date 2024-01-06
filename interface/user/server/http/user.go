package http

import (
	"context"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strconv"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
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
		response.Fail(c, "参数验证失败", nil)
		return
	}
	// 正则表达式匹配邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		response.Fail(c, "邮箱格式不正确", nil)
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

	token, err := utils.GenerateToken(resp.UserId, resp.Email)
	if err != nil {
		logger.Error("failed to generate user token", zap.Error(err))
		response.Fail(c, err.Error(), nil)
		return
	}

	response.Success(c, "登录成功", gin.H{"token": token, "user_info": resp})
}

type RegisterRequest struct {
	//Email    string `json:"email" binding:"required,email"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	//ConfirmPass string `json:"confirm_password" binding:"required,eqfield=Password"`
	ConfirmPass string `json:"confirm_password" binding:"required"`
	Avatar      string `json:"avatar"`
	Nickname    string `json:"nickname"`
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
		response.Fail(c, "参数验证失败", nil)
		return
	}
	if req.Password != req.ConfirmPass {
		response.Fail(c, "密码和确认密码不匹配", nil)
		return
	}
	// 正则表达式匹配邮箱格式
	emailRegex := regexp2.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, 0)
	if isMatch, _ := emailRegex.MatchString(req.Email); !isMatch {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "邮箱格式不正确"})
		return
	}

	//最少包括一个数字，大小字符，最短8个字符，最长20个字符
	emailRegex = regexp2.MustCompile(`^(?=.*[0-9])(?=.*[a-zA-Z]).{6,50}$`, 0)
	if isMatch, _ := emailRegex.MatchString(req.Password); !isMatch {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "密码格式不正确"})
		return
	}
	if isMatch, _ := emailRegex.MatchString(req.ConfirmPass); !isMatch {
		response.Fail(c, "密码格式不正确", nil)
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
		c.Error(err)
		return
	}

	response.Success(c, "注册成功", gin.H{"user_id": resp.UserId})
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
	// 正则表达式匹配邮箱格式
	emailRegex := regexp2.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, 0)
	if isMatch, _ := emailRegex.MatchString(email); !isMatch {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "邮箱格式不正确"})
		return
	}
	switch GetType(gtype) {
	case EmailType:
		resp, err := userClient.GetUserInfoByEmail(context.Background(), &user.GetUserInfoByEmailRequest{
			Email: email,
		})
		if err != nil {
			c.Error(err)
			return
		}
		if resp == nil {
			response.Fail(c, "用户不存在", nil)
			return
		}
		response.Success(c, "查询成功", gin.H{"user_info": resp})
	case UserIdType:
		resp, err := userClient.UserInfo(context.Background(), &user.UserInfoRequest{
			UserId: userId,
		})
		if err != nil {
			c.Error(err)
		}
		if resp == nil {
			response.Fail(c, "用户不存在", nil)
			return
		}
		response.Success(c, "查询成功", gin.H{"user_info": resp})
	}
}
