package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/encryption"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	resp "github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

// EncryptionMiddleware 加密和解密中间件
func EncryptionMiddleware(encryptor encryption.Encryptor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求方法是否为 GET
		if c.Request.Method != http.MethodGet && encryptor.IsEnable() && c.Request.URL.Path != "/api/v1/user/system/key/get" {
			request := new(encryption.SecretResponse)
			if err := c.ShouldBindJSON(&request); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "msg": "非加密请求体"})
				return
			}
			if request.Secret == "" || request.Message == "" {
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "msg": "参数验证失败"})
				return
			}
			// 进行解密操作
			key, err := encryptor.DecryptMessage(request.Secret)
			if err != nil {
				fmt.Println("解析对称秘钥失败：", err)
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "msg": "Failed to read request body"})
				return
			}

			data, err := helper.DecryptMessageWithPassword([]byte(key), request.Message)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "msg": "Failed to read request body"})
				return
			}
			fmt.Println("解密后消息 =>", data)
			//将解密后的数据放入上下文，供后续处理使用
			c.Request.Body = ioutil.NopCloser(strings.NewReader(data))
		}
		c.Next()
		if storedResponse, exists := c.Get("response"); exists {
			response := storedResponse.(resp.BaseResponse)
			msgStr, err := json.Marshal(response)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  http.StatusText(http.StatusInternalServerError),
				})
				return
			}
			if !encryptor.IsEnable() || c.Request.URL.Path == "/api/v1/user/system/key/get" || response.Code != 200 {
				c.JSON(http.StatusOK, gin.H{"code": response.Code, "msg": response.Msg, "data": response.Data})
				return
			}
			rkey, err := encryption.GenerateRandomKey(32)
			if err != nil {
				fmt.Println("生成秘钥失败:", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  http.StatusText(http.StatusInternalServerError),
				})
				return
			}
			conn, err := db.NewDefaultMysqlConn().GetConnection()
			if err != nil {
				fmt.Println("db conn failed", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  http.StatusText(http.StatusInternalServerError),
				})
				return
			}
			ea := encryption.NewEncryptedAuthenticator(conn)
			if err != nil {
				fmt.Println("init db", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  http.StatusText(http.StatusInternalServerError),
				})
				return
			}
			var userId string
			if userIdResponse, exists := c.Get("user_id"); exists {
				id := userIdResponse.(string)
				if id != "" {
					userId = id
				}
			} else {
				userId, err = pkghttp.ParseTokenReUid(c)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code": 401,
						"msg":  err.Error(),
					})
					return
				}
			}
			if userId == "" {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "响应用户ID为空",
				})
			}
			user, err := ea.QueryUser(userId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "用户不存在",
				})
				return
			}
			if user.PublicKey == "" {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "用户未设置公钥,加密失败",
				})
				return
			}
			// 进行加密操作
			encryptedResponse, err := encryptor.SecretMessage(string(msgStr), user.PublicKey, []byte(rkey))
			if err != nil {
				fmt.Println("加密失败：", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "PublicKey error:" + err.Error(),
				})
				return
			}
			fmt.Println("响应消息 =>", encryptedResponse)
			c.JSON(http.StatusOK, gin.H{"message": encryptedResponse.Message, "secret": encryptedResponse.Secret})

		}
		return
	}
}
