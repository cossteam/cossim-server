package utils

//type Claims struct {
//	UserId string `json:"user_id"`
//	Email  string `json:"email"`
//	jwt.RegisteredClaims
//}
//
//// 验证token是否有效
//func VerifyLogin(tokenStr string) (*entity.User, error) {
//	if len(tokenStr) == 0 {
//		return nil, errors.New("")
//	}
//	// 缓存解析后的 token 和 claims
//	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
//		return []byte(config.JwtSecretKey), nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	claims, ok := token.Claims.(*Claims)
//	if !ok {
//		return nil, errors.New("invalid claims")
//	}
//	if claims.UserId == "" {
//		return nil, errors.New("")
//	}
//	user, err := persistence.UserR.GetUserInfoByUid(claims.UserId)
//	if err != nil {
//		return nil, errors.New("用户不存在")
//	}
//
//	return user, nil
//}
//
//// 生成 token
//func GenerateToken(user *entity.User, refresh bool) string {
//	var token string
//
//	days := 30 //token过期时间
//	if refresh {
//		//刷新时间
//		t := jwt.NewWithClaims(jwt.SigningMethodHS256,
//			Claims{
//				UserId:           user.ID,
//				Email:            user.Email,
//				RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Duration(days) * 24 * time.Hour)}},
//			})
//		tokenStr, err := t.SignedString([]byte(config.JwtSecretKey))
//		if err != nil {
//			return ""
//		}
//		token = tokenStr
//	} else {
//		//todo 从redis查询token
//		//core.DB.First(&token, user.ID)
//		//user.Token = token
//	}
//	return token
//}
