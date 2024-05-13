package entity

type AuthClaims struct {
	UserID   string
	Email    string
	DriverID string
}

type UserToken struct {
	Token string
}
