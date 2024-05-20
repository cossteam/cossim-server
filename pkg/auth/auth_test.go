package auth_test

import (
	"testing"
)

func TestValidateToken(t *testing.T) {

	//connection, err := db.NewMySQLFromDSN("root:888888@tcp(127.0.0.1:33066)/coss?allowNativePasswords=true&timeout=800ms&readTimeout=200ms&writeTimeout=800ms&parseTime=true&loc=Local&charset=utf8,utf8mb4").GetConnection()
	//if err != nil {
	//	panic(err)
	//}
	//
	//authenticator := auth.NewAuthenticator(connection)
	//
	//// 模拟的claims数据
	//claims := &auth.Claims{
	//	ID: "123123",
	//	Email:  "test@example.com",
	//}
	//
	//token, err := utils.GenerateToken(claims.ID, claims.Email)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// 测试验证token
	//result, err := authenticator.ValidateToken(token)
	//assert.True(t, result)
	//assert.NoError(t, err)
}
