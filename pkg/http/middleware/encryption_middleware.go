package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/encryption"
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
		if c.Request.Method != http.MethodGet && encryptor.IsEnable() && c.Request.URL.Path != "/api/v1/user/system/public_key" {
			request := new(encryption.SecretResponse)
			if err := c.ShouldBindJSON(&request); err != nil {
				fmt.Println("解析消息失败：", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "非加密请求体"})
				return
			}
			if request.Secret == "" || request.Message == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "参数验证失败"})
				return
			}

			key, err := encryptor.DecryptMessage(request.Secret)
			if err != nil {
				fmt.Println("解析对称秘钥失败：", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "Failed to read request body"})
				return
			}

			data, err := helper.DecryptMessageWithPassword([]byte(key), request.Message)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "Failed to read request body"})
				return
			}
			fmt.Println("解密后消息 =>", data)
			// 将解密后的数据放入上下文，供后续处理使用
			c.Request.Body = ioutil.NopCloser(strings.NewReader(data))
		}
		c.Next()
		if storedResponse, exists := c.Get("response"); exists {
			response := storedResponse.(resp.BaseResponse)
			msgStr, err := json.Marshal(response)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": http.StatusText(http.StatusInternalServerError)})
				return
			}

			//不需要加密的情况
			if !encryptor.IsEnable() || c.Request.URL.Path == "/api/v1/user/system/public_key" || response.Code != 200 {
				c.JSON(http.StatusOK, gin.H{"code": response.Code, "msg": response.Msg, "data": response.Data})
				return
			}

			rkey, err := encryption.GenerateRandomKey(32)
			if err != nil {
				fmt.Println("生成秘钥失败:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": http.StatusText(http.StatusInternalServerError)})
				return
			}

			if c.Value(constants.UserID) == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "响应用户ID为空"})
				return
			}
			userId := c.Value(constants.UserID).(string)
			if userId == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "响应用户ID为空"})
				return
			}

			if c.Value(constants.PublicKey) == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "响应用户公钥为空"})
				return
			}
			publicKey := c.Value(constants.PublicKey).(string)
			if publicKey == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "用户未设置公钥,加密失败"})
				return
			}

			// 并发处理加密响应
			encryptedResponse, err := encryptor.SecretMessage(string(msgStr), publicKey, []byte(rkey))
			if err != nil {
				fmt.Println("加密失败：", err)
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "PublicKey error:" + err.Error()})
				return
			}
			fmt.Println("响应消息 =>", encryptedResponse)
			c.JSON(http.StatusOK, gin.H{"message": encryptedResponse.Message, "secret": encryptedResponse.Secret})

		}
	}
}
