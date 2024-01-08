package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/encryption"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

// EncryptionMiddleware 加密和解密中间件
func EncryptionMiddleware(encryptor encryption.Encryptor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求方法是否为 GET
		if c.Request.Method != http.MethodGet && encryptor.IsEnable() && c.Request.URL.Path != "/api/v1/user/key/set" {
			var request encryption.SecretResponse
			if err := c.ShouldBindJSON(&request); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "error": "Failed to read request body"})
				return
			}
			// 进行解密操作
			key, err := encryptor.DecryptMessage(request.Secret)
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "error": "Failed to read request body"})
				return
			}

			data, err := encryptor.DecryptMessageWithKey(request.Message, key)
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "error": "Failed to read request body"})
				return
			}
			//将解密后的数据放入上下文，供后续处理使用
			c.Request.Body = ioutil.NopCloser(strings.NewReader(data))
		}
		c.Next()

		if storedResponse, exists := c.Get("response"); exists {
			response := storedResponse.(utils.Response)
			msgStr, err := json.Marshal(response)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  http.StatusText(http.StatusInternalServerError),
				})
				return
			}
			if !encryptor.IsEnable() {
				c.String(http.StatusOK, string(msgStr))
				return
			}
			rkey, err := encryptor.GenerateRandomKey(32)
			if err != nil {
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
			thisId, err := pkghttp.ParseTokenReUid(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 401,
					"msg":  err.Error(),
				})
				return
			}
			user, err := ea.QueryUser(thisId)
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
			encryptedResponse, err := encryptor.SecretMessage(string(msgStr), user.PublicKey, rkey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "PublicKey error:" + err.Error(),
				})
				return
			}

			// 替换响应体为加密后的数据
			msg, err := json.Marshal(encryptedResponse)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  http.StatusText(http.StatusInternalServerError),
				})
				return
			}
			c.String(http.StatusOK, string(msg))
		}

		return
	}
}
