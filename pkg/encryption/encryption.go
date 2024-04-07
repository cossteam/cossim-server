package encryption

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"gorm.io/gorm"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
)

// Encryptor 接口定义了加密和解密的方法
type Encryptor interface {
	GenerateKeyPair() error
	SecretMessage(message string, publicKey string, rkey []byte) (*SecretResponse, error)
	DecryptMessage(message string) (string, error)
	DecryptMessageWithKey(key string, message string) (string, error)
	GetPrivateKey() string
	GetPublicKey() string
	IsEnable() bool
	SetPublicKey(publicKey string)
	SetPrivateKey(privateKey string)
	ReadKeyPair() error
	GetSecretMessage(message string, userID string) (string, error)
	QueryUser(userID string) (entity.User, error)
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
	db         *gorm.DB
}

// 生成随机对称秘钥
func GenerateRandomKey(length int) (string, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := big.NewInt(int64(len(charset)))

	var result strings.Builder
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			panic(err)
		}
		result.WriteByte(charset[randomIndex.Int64()])
	}

	return result.String(), nil
}

// 加密消息响应格式
type SecretResponse struct {
	Message string `json:"message"`
	Secret  string `json:"secret"`
}

// NewEncryptor 创建一个新的 Encryptor 实例
func NewEncryptor(passphrase []byte, name, email string, rsaBits int, enable bool, db *gorm.DB) Encryptor {
	return &MyEncryptor{
		passphrase: passphrase,
		name:       name,
		email:      email,
		rsaBits:    rsaBits,
		enable:     enable,
		db:         db,
	}
}

//type EncryptedAuthenticator struct {
//	DB *gorm.DB
//}
//
//func NewEncryptedAuthenticator(db *gorm.DB) *EncryptedAuthenticator {
//	return &EncryptedAuthenticator{
//		DB: db,
//	}
//}

const _queryUser = "SELECT * FROM users WHERE id = ?"

// QueryUser retrieves user information by user ID
func (e *MyEncryptor) QueryUser(userID string) (entity.User, error) {
	var user entity.User
	if err := e.db.Raw(_queryUser, userID).Scan(&user).Error; err != nil {
		return entity.User{}, err
	}
	return user, nil
}

// 根据用户id返回加密后消息
func (e *MyEncryptor) GetSecretMessage(message string, userID string) (string, error) {
	if !e.enable {
		return message, nil
	}

	userInfo, err := e.QueryUser(userID)
	if err != nil {
		return message, err
	}
	if userInfo.PublicKey == "" {
		return message, fmt.Errorf("publicKey is nil")
	}
	rkey, err := GenerateRandomKey(32)
	if err != nil {
		return message, err
	}

	data := new(SecretResponse)

	armor, err := helper.EncryptMessageWithPassword([]byte(rkey), message)
	if err != nil {
		return message, err
	}
	data.Message = armor
	marmor, err := helper.EncryptMessageArmored(userInfo.PublicKey, rkey)
	if err != nil {
		return message, err
	}
	data.Secret = marmor

	marshal, err := json.Marshal(data)
	if err != nil {
		return message, err
	}

	return string(marshal), nil
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

	cacheDir := ".cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err := os.Mkdir(cacheDir, 0755) // 创建文件夹并设置权限
		if err != nil {
			return err
		}
	}
	// 保存私钥到文件
	privateKeyFile, err := os.Create(cacheDir + "/private_key")
	if err != nil {
		return err
	}

	_, err = privateKeyFile.WriteString(rsaKey)
	if err != nil {
		privateKeyFile.Close()
		return err
	}
	privateKeyFile.Close()

	// 保存公钥到文件
	publicKeyFile, err := os.Create(cacheDir + "/public_key")
	if err != nil {
		return err
	}
	keyRing, err := crypto.NewKeyFromArmoredReader(strings.NewReader(rsaKey))
	if err != nil {
		return err
	}

	publicKey, err := keyRing.GetArmoredPublicKey()
	if err != nil {
		return err
	}
	_, err = publicKeyFile.WriteString(publicKey)
	if err != nil {
		publicKeyFile.Close()
		return err
	}
	publicKeyFile.Close()
	e.publicKey = publicKey
	return nil
}
func (e *MyEncryptor) ReadKeyPair() error {
	cacheDir := ".cache"

	//getwd, err := os.Getwd()
	//if err != nil {
	//	return err
	//}
	//
	////输出当前工作目录
	//fmt.Println("ReadKeyPair pwd => ", getwd)

	// 检查目录是否存在
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return errors.New("缓存目录不存在")
	}

	// 读取私钥文件
	privateKeyFile, err := os.Open(cacheDir + "/private_key")
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyBytes, err := ioutil.ReadAll(privateKeyFile)
	if err != nil {
		return err
	}
	e.privateKey = string(privateKeyBytes)

	// 读取公钥文件
	publicKeyFile, err := os.Open(cacheDir + "/public_key")
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := ioutil.ReadAll(publicKeyFile)
	if err != nil {
		return err
	}
	e.publicKey = string(publicKeyBytes)

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
func (e *MyEncryptor) DecryptMessageWithKey(key string, message string) (string, error) {
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

func (e *MyEncryptor) SetPrivateKey(privateKey string) {
	e.privateKey = privateKey
}
func (e *MyEncryptor) SetPublicKey(publicKey string) {
	e.publicKey = publicKey
}

func (e *MyEncryptor) SetEnable(en bool) {
	e.enable = en
}
