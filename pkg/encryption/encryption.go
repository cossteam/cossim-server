package encryption

import (
	"crypto/rand"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/cossim/coss-server/service/user/domain/entity"
	"gorm.io/gorm"
	"strings"
)

// Encryptor 接口定义了加密和解密的方法
type Encryptor interface {
	GenerateRandomKey(keySize int) ([]byte, error)
	GenerateKeyPair() error
	SecretMessage(message string, publicKey string, rkey []byte) (*SecretResponse, error)
	DecryptMessage(message string) (string, error)
	DecryptMessageWithKey(message, key string) (string, error)
	GetPrivateKey() string
	GetPublicKey() string
	IsEnable() bool
}

// MyEncryptor 结构体实现了 Encryptor 接口
type MyEncryptor struct {
	enable     bool
	privateKey string
	publicKey  string
	passphrase []byte
	name       string
	email      string
	rsaBits    int
}

// 生成随机对称秘钥
func GenerateRandomKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// 加密消息响应格式
type SecretResponse struct {
	Message string `json:"message"`
	Secret  string `json:"secret"`
}

// NewEncryptor 创建一个新的 Encryptor 实例
func NewEncryptor(passphrase []byte, name, email string, rsaBits int, enable bool) Encryptor {
	return &MyEncryptor{
		passphrase: passphrase,
		name:       name,
		email:      email,
		rsaBits:    rsaBits,
		enable:     enable,
	}
}

type EncryptedAuthenticator struct {
	DB *gorm.DB
}

func NewEncryptedAuthenticator(db *gorm.DB) *EncryptedAuthenticator {
	return &EncryptedAuthenticator{
		DB: db,
	}
}

const _queryUser = "SELECT * FROM users WHERE id = ?"

// QueryUser retrieves user information by user ID
func (a *EncryptedAuthenticator) QueryUser(userID string) (entity.User, error) {
	var user entity.User
	if err := a.DB.Raw(_queryUser, userID).Scan(&user).Error; err != nil {
		return entity.User{}, err
	}
	return user, nil
}

// GenerateRandomKey 生成随机对称密钥
func (e *MyEncryptor) GenerateRandomKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// SecretMessage 根据公钥与随机对称秘钥加密消息
func (e *MyEncryptor) SecretMessage(message string, publicKey string, rkey []byte) (*SecretResponse, error) {
	if publicKey == "" {
		return nil, fmt.Errorf("public key is empty")
	}
	if rkey == nil {
		return nil, fmt.Errorf("random key is empty")
	}
	data := new(SecretResponse)
	armor, err := helper.EncryptMessageWithPassword(rkey, message)
	if err != nil {
		return nil, err
	}
	data.Message = armor

	marmor, err := helper.EncryptMessageArmored(publicKey, string(rkey))
	if err != nil {
		return nil, err
	}
	data.Secret = marmor

	return data, nil
}

// GenerateKeyPair 使用个人信息生成公私钥
func (e *MyEncryptor) GenerateKeyPair() error {
	rsaKey, err := helper.GenerateKey(e.name, e.email, e.passphrase, "rsa", e.rsaBits)
	if err != nil {
		return err
	}
	// 保存私钥到结构体
	e.privateKey = rsaKey

	keyRing, err := crypto.NewKeyFromArmoredReader(strings.NewReader(rsaKey))
	if err != nil {
		return err
	}

	publicKey, err := keyRing.GetArmoredPublicKey()
	if err != nil {
		return err
	}
	e.publicKey = publicKey
	return nil
}

// DecryptMessage 使用私钥解密消息
func (e *MyEncryptor) DecryptMessage(message string) (string, error) {
	decrypted, err := helper.DecryptBinaryMessageArmored(e.privateKey, e.passphrase, message)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// DecryptMessageWithKey 使用对称密钥解密消息
func (e *MyEncryptor) DecryptMessageWithKey(message, key string) (string, error) {
	decrypted, err := helper.DecryptMessageWithPassword([]byte(key), message)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

// 获取私钥
func (e *MyEncryptor) GetPrivateKey() string {
	return e.privateKey
}

// 获取公钥
func (e *MyEncryptor) GetPublicKey() string {
	return e.publicKey
}

func (e *MyEncryptor) IsEnable() bool {
	return e.enable
}
