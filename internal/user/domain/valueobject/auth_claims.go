package valueobject

import "errors"

func NewAuthClaims(UserID, Email, DriverID, PublicKey string) (AuthClaims, error) {
	if UserID == "" || Email == "" || DriverID == "" || PublicKey == "" {
		return AuthClaims{}, errors.New("userID, email, driverID, publicKey is required")
	}
	return AuthClaims{
		userID:    UserID,
		email:     Email,
		driverID:  DriverID,
		publicKey: PublicKey,
	}, nil
}

type AuthClaims struct {
	userID    string
	email     string
	driverID  string
	publicKey string
}

func (v AuthClaims) UserID() string {
	return v.userID
}

func (v AuthClaims) Email() string {
	return v.email
}

func (v AuthClaims) DriverID() string {
	return v.driverID
}

func (v AuthClaims) PublicKey() string {
	return v.publicKey
}
