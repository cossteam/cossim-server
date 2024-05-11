package entity

type AuthClaims struct {
	UserID    string
	Email     string
	DriverID  string
	PublicKey string
}

type UserToken struct {
	Token string
}
