package remote

import (
	"github.com/cossim/coss-server/pkg/email"
)

type SmtpService interface {
	GenerateEmailVerificationContent(addr string, userID string, code string) string
	SendEmail(Email, subject, body string) error
}

var _ SmtpService = &smtpService{}

type smtpService struct {
	client email.EmailProvider
}

func (s *smtpService) GenerateEmailVerificationContent(addr string, userID string, code string) string {
	return s.client.GenerateEmailVerificationContent(addr, userID, code)
}

func NewSmtpService(client email.EmailProvider) SmtpService {
	return &smtpService{client: client}
}

func (s *smtpService) SendEmail(Email, subject, body string) error {
	return s.client.SendEmail(Email, subject, body)
}
