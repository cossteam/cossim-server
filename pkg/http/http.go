package http

// ParseTokenReUid 解析请求头中的 token 返回 uid
//func ParseTokenReUid(ctx *gin.Context) (string, error) {
//	tokenString := ctx.GetHeader("Authorization")
//
//	if tokenString == "" {
//		return "", errors.New("authorization header is empty")
//	}
//
//	token := tokenString[7:]
//	if token != "" {
//		_, c2, err := utils.ParseToken(token)
//		if err != nil {
//			return "", err
//		}
//		return c2.ID, nil
//	}
//
//	return "", errors.New("token is empty")
//}

// ParseTokenReDriverId 解析请求头中的 token 返回 设备id
//func ParseTokenReDriverId(ctx *gin.Context) (string, error) {
//	tokenString := ctx.GetHeader("Authorization")
//
//	if tokenString == "" {
//		return "", errors.New("authorization header is empty")
//	}
//
//	token := tokenString[7:]
//	if token != "" {
//		_, c2, err := utils.ParseToken(token)
//		if err != nil {
//			return "", err
//		}
//		return c2.DriverID, nil
//	}
//
//	return "", errors.New("token is empty")
//}
