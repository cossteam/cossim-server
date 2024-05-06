package smtp

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/email"
	"github.com/wneessen/go-mail"
	"log"
	"net/url"
)

type Storage struct {
	c         *mail.Client
	thisEmail string
}

func NewSmtpStorage(server string, port int, username string, password string) (email.EmailProvider, error) {
	obj := &Storage{
		thisEmail: username,
	}
	//cc, err := mail.NewClient("smtp.qq.com", mail.WithPort(25), mail.WithSMTPAuth(mail.SMTPAuthPlain),
	//	mail.WithUsername("2318266924@qq.com"), mail.WithPassword("zjnudhwoiuknecgh"))
	//if err != nil {
	//	log.Fatalf("failed to create mail client: %s", err)
	//}
	cc, err := mail.NewClient(server, mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username), mail.WithPassword(password))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}
	obj.c = cc
	return obj, err
}

//func TestEncryption(t *testing.T) {
//	m := mail.NewMsg()
//	if err := m.From("2318266924@qq.com"); err != nil {
//		log.Fatalf("failed to set From address: %s", err)
//	}
//	if err := m.To("2622788078@qq.com"); err != nil {
//		log.Fatalf("failed to set To address: %s", err)
//	}
//	m.Subject("老铁拉屎没纸")
//	m.SetBodyString(mail.TypeTextPlain, "Do you like this mail? I certainly do!")
//
//
//	if err := c.DialAndSend(m); err != nil {
//		log.Fatalf("failed to send mail: %s", err)
//	}
//}

func (s Storage) SendEmail(to, subject, body string) error {
	m := mail.NewMsg()
	if err := m.From(s.thisEmail); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, body)

	if err := s.c.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func (s Storage) GenerateEmailVerificationContent(gatewayAdd string, userId string, key string) string {
	baseURL := gatewayAdd + "/api/v1/user/activate" // 替换成你的网站的基本URL
	verifyURL := fmt.Sprintf("%s?user_id=%s&key=%s", baseURL, userId, key)

	// 将验证链接进行URL编码
	encodedURL := url.QueryEscape(verifyURL)

	content := fmt.Sprintf("激活请点击：<a href=\"%s\">%s</a>", encodedURL, verifyURL)

	return content
}
