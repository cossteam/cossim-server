package email

type EmailProvider interface {
	SendEmail(to, subject, body string) error
	//SendVerificationEmail(to, subject, body, token string) error
	GenerateEmailVerificationContent(addr string, userId string, key string) string
}
