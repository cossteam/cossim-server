package encryption_test

import (
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/encryption/config"
	"testing"
)

func TestEncryption(t *testing.T) {
	readString, err := encryption.GenerateRandomKey(32)
	if err != nil {
		panic(err)
	}
	err = config.Init()
	if err != nil {
		return
	}
	en := encryption.NewEncryptor([]byte(config.Conf.Encryption.Passphrase), config.Conf.Encryption.Name, config.Conf.Encryption.Email, config.Conf.Encryption.RsaBits)
	err = en.GenerateKeyPair()
	if err != nil {
		return
	}
	resp, err := en.SecretMessage("{\"id\":\"666666\"}", en.GetPublicKey(), readString)
	fmt.Println("公钥:", en.GetPublicKey())
	if err != nil {
		return
	}
	j, err := json.Marshal(resp)
	fmt.Println("加密后消息：", string(j))

	if en.GetPrivateKey() != "" {
		key, _ := en.DecryptMessage(resp.Secret)
		msg, _ := en.DecryptMessageWithKey(resp.Message, key)
		fmt.Println("解密后消息：", msg)
	}
}
