package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

// EncryptionMiddleware 加密和解密中间件
func EncryptionMiddleware(encryptor encryption.Encryptor) gin.HandlerFunc {
	return func(c *gin.Context) {
		var response encryption.SecretResponse
		if err := c.ShouldBindJSON(&response); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
		//把加密数据解析结构体
		fmt.Println("开始解密秘钥", response.Secret)
		// 进行解密操作
		key, err := encryptor.DecryptMessage(response.Secret)
		fmt.Println("1111111111111111111111111111111112222")
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
		fmt.Println("key:", key)

		data, err := encryptor.DecryptMessageWithKey(response.Message, key)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
		fmt.Println("解密后:", data)

		//将解密后的数据放入上下文，供后续处理使用
		//c.Set("decryptedData", data)
		c.Request.Body = ioutil.NopCloser(strings.NewReader(data))
		// 继续处理请求
		c.Next()

		responseData, exists := c.Get("responseData")
		if !exists {
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
		// 进行加密操作
		encryptedResponse, err := encryptor.SecretMessage(string(responseData.([]byte)), encryptor.GetPublicKey(), rkey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		// 替换响应体为加密后的数据
		c.Writer.WriteHeader(200)
		c.Writer.Write([]byte(encryptedResponse.Secret))
	}
}
