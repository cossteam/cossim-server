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
	fmt.Println("随机秘钥:\n", readString)
	err = config.Init()
	if err != nil {
		return
	}
	en := encryption.NewEncryptor([]byte(config.Conf.Encryption.Passphrase), config.Conf.Encryption.Name, config.Conf.Encryption.Email, config.Conf.Encryption.RsaBits, config.Conf.Encryption.Enable)
	//resp := encryption.SecretResponse{
	//	Message: "-----BEGIN PGP MESSAGE-----\n\n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\ncOM3KN2aqKBLS5LwxSW4khXwKjtJMyPeHYI=\n=uOTk\n-----END PGP MESSAGE-----",
	//	Secret: "-----BEGIN PGP MESSAGE-----\n\nwcBMAwwUZOchgEUSAQgAn4To8KYC+/aMsagHU2Xhw87TboadMvsp4lkvqJ8M\nZRmM/FCv+l/8uYWFbT21r1EEvqpUjDZ/wbjQvslPR4gRdCQb5MfqAlhNR8Ek\nsfFnuoR4JDSyvyRl99PWsJT08ARgM3JqYdQQgPjByFDOyOX0dpeaiNeHheFY\ntHuB3NW5gJM4TTdPTyA8K0aw8mJhVDL5keiHbqkBsFZUzsy/loiWwl2h+aSi\n1McBzvUf5pNihMauMTV1Ei/vAc/izCzwtzjV54+kwi2URo7eGKc8NpVKBz8n\nyoo3S153lub8ZKKwLXZJj46stidU/H0NTmh9bPrC4jMuOU/I6Hb6wuZzZGO7\nYdI9ASUFQn0mMuuSPjJ5gG/YNDOF0ypPqP4CNUH6ibkYHzKLn3xbHUXk60+b\nzgtFWPMVqPOJFel8eG/Yciaslg==\n=eDPI\n-----END PGP MESSAGE-----"
	//}
	//err = en.GenerateKeyPair()
	//if err != nil {
	//	return
	//}
	key := "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nxsBNBGWc+okBCAD3kT3bXBYTTU6QEUvQhf4/yZ/dhaake3RWzQ1FdWcgVDl+\noQyBjbIU2h4O3hTuCLP+8erR79PEyIorOLo75vCRppcy0n8paqrMB3rTXhkq\njLSLvr6K2RDXM1OkYQ+xKyt1A75ksTKL6Y8mrEp57jQYb4fDMqTPQlNGvmzN\n9Q9Pp/vQ3bpVUSgU7iCYJ5JzH5muie7xOl+n7jx/pwHXV8Jq2yaQirMZIfsl\n7AqsWAKNiecouJT0/LXVcdfQ2oIi1+Y3DQuoTZHpRWEB4nAWdX24x5CPbjcq\n+WdDZG36CDAlbzVcJuquENlmGresHiJpLNmK2ocS1DB125s+PJZLEQB5ABEB\nAAHNFkNPU1MgPGNvc3NAY29zc2ltLmNvbT7CwIoEEAEIAD4FgmWc+okECwkH\nCAmQSj3eynAecCADFQgKBBYAAgECGQECmwMCHgEWIQR8Uhlb8e4M6qqLfINK\nPd7KcB5wIAAAiiUH/29QKqTsogC2aJz6vAh5WO2lVJ6Z8e1QAzreMVaLrh+u\n6swdHbeMPx8hp8td1eEQYqX2V4pW8w5gm3GcdJxp5XAsw/9Kryu24N0c2SZg\nxJVbVIasWSlIfuj9cGRnavX6RrRKh3P4s8L6MOlfH4UU2ypo0pMPBPy/L5XC\nLtX7T8a255fz0D569Q+1G4bb6xIPBr5ark02OCHRGWDrYUWvtvBnJAbqGrYY\nRo4BrYY9nob2+IkXTKc/gvguj7yLLGfS4y7T1SoJC1bpQx73RpiAVuO/NPiQ\nxzJJ92mLMFwX8PyZTMSo7gOEn3PVwtBtae2WiTUKzpjnYGbl83HjgXQuYWrO\nwE0EZZz6iQEIAPB1XijJloMvT06N0U0lqpfzCmNjmlpa66liabkQDPsoSJOu\nqgGWKG0xIM8aiurh+aNdBA4bCJGVu7Hrzi9lDx7KLAf8EJlaRhoLHvUWof63\nvB9h0SOZcbt2eUz1kFj1H+x/FyR+AkKXsGLqWyYUvpQtvnr0FxtK9zGG5aUD\nnGMbYcZPklnWFTwA07G1OnusD9jl8IBUeMST/hNPP3CnIGgSfo7AnLyRhHyx\nGdeX92nJPNqowXXTTBgXr5dx7ba2APjv+DF1VuoUBnOKdouiVEzdGhIpFkAU\nkpIW/n3jLJwZZukopPVQTf8bQReM0fqYJfzdCqU0eJuBo+4usvnwn7kAEQEA\nAcLAdgQYAQgAKgWCZZz6iQmQSj3eynAecCACmwwWIQR8Uhlb8e4M6qqLfINK\nPd7KcB5wIAAASKcIAMQDduQXh0c7wgTpYsSAEzlL+a/2rmn1vL/vrfXvmQ+e\nFBD+zuUEru/Yg9QqSbnYjqLKwcJSRDnfTnh/IWo/PKkusJv/TsfAs8sC01dt\nqxuAlnvt5WE9UGbiSXnjPUdHfNbhFPaEkmG83mD6nHC7RpQxJqSKWm+ObCCC\n016lxwA1KCYxQmJ/cKnESBu2AvVcvDPctVGSYgTN2nSbYUX5CDqjwVrc1z+I\nWOLX/0IVmplamiq9ZEf//xjNRPoxrgIeh2YWkFB9BQJwtpca/FU8ciprvdNQ\nC8tzYYH4+6B503aWQ8X31aST76+m3O1iQUe8KSeqrGFl0tzcI84BfsLN14k=\n=LXB9\n-----END PGP PUBLIC KEY BLOCK-----"
	message, err := en.SecretMessage("{\"code\":200,\"msg\":\"登录成功\",\"data\":{\"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiN2Q5NTIwZGQtYmNlNC00MWRlLWFmNTAtMTIxNTY1NDBjNmIyIiwiZW1haWwiOiIxMjNAcXEuY29tIiwiZXhwIjoxNzA3Mzc3Mzk4fQ._9QmP78ZlUuH2jsSa-BBw2ck0VRnUwfvbPh5dN7qXB8\",\"user_info\":{\"UserId\":\"7d9520dd-bce4-41de-af50-12156540c6b2\",\"Email\":\"123@qq.com\"}}}", key, []byte(readString))
	if err != nil {
		fmt.Println(err)
		return
	}

	j, err := json.Marshal(message)

	fmt.Println("消息:", string(j))

	//resp, err := en.SecretMessage("{\n  \"message\": \"-----BEGIN PGP MESSAGE-----\\n\\nwy4ECQMIoL/bAHnpmd/gZP6lR1LTZGHUx3iYSKK1rxB1J1UzM+PtVtnd0lNJ\\neozv0jQBPzE9IBiwCNOhPlFfGCpj60aDrDyCHq49m7BPpggZ7E0aEFKKd64a\\nuCGBLBZ2knRmp6we\\n=ddUS\\n-----END PGP MESSAGE-----\\n\",\n  \"secret\": \"-----BEGIN PGP MESSAGE-----\\n\\nwcBMA5st/PWcI/vwAQf+MwoisWX1TCjM3HOC0aGWcKQV1Ufuw6JgRTWVSANy\\ni7NgVG09lqKkRB5ZzgPC0W+pTIUNXeJyTCjgR/VgGMG8S+PftDrm+mjFb0R2\\nGYkhDooPNIwhRFCkYQL8nQL1qSw7gzorXxn2STuLgZJkxi+UZTCBCXJ9pb8O\\nsmsQ6OtMbh362JoGr0FHpZSgBKHul2hIhv1gRFzXG+uvhnO+YNptSpqTOZAl\\nuyST8/+oObn5Dt30ze3zeo0wvLX+11zRXo2DOZOVeqGJ6ZkMKUS2lUeEUbRT\\nbFY6HGFe0zdVRMgl37FIfNXocJ5UhwOOCGHjA2S5ouPWWeEmre1WNYyBCCQ9\\nQ9I9ATQ32EHwPqwj96vx6s58hNCgrkH9rMSFfuO/aOWoRFUUe1EQWXD+eG+1\\neNRV/MkfFLHHB/7vV+Dk7INoqw==\\n=ua5l\\n-----END PGP MESSAGE-----\\n\"\n}", en.GetPublicKey(), readString)
	//fmt.Println("公钥:", en.GetPublicKey())
	//if err != nil {
	//	return
	//}
	//j, err := json.Marshal(resp)
	//fmt.Println("加密后消息：", string(j))
	//en.SetPrivateKey("-----BEGIN PGP PRIVATE KEY BLOCK-----\n\n\n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\n=0alX\n-----END PGP PRIVATE KEY BLOCK-----")
	//if en.GetPrivateKey() != "" {
	//	key, err := en.DecryptMessage("-----BEGIN PGP MESSAGE-----\\n\\nwcBMA5st/PWcI/vwAQf+MwoisWX1TCjM3HOC0aGWcKQV1Ufuw6JgRTWVSANy\\ni7NgVG09lqKkRB5ZzgPC0W+pTIUNXeJyTCjgR/VgGMG8S+PftDrm+mjFb0R2\\nGYkhDooPNIwhRFCkYQL8nQL1qSw7gzorXxn2STuLgZJkxi+UZTCBCXJ9pb8O\\nsmsQ6OtMbh362JoGr0FHpZSgBKHul2hIhv1gRFzXG+uvhnO+YNptSpqTOZAl\\nuyST8/+oObn5Dt30ze3zeo0wvLX+11zRXo2DOZOVeqGJ6ZkMKUS2lUeEUbRT\\nbFY6HGFe0zdVRMgl37FIfNXocJ5UhwOOCGHjA2S5ouPWWeEmre1WNYyBCCQ9\\nQ9I9ATQ32EHwPqwj96vx6s58hNCgrkH9rMSFfuO/aOWoRFUUe1EQWXD+eG+1\\neNRV/MkfFLHHB/7vV+Dk7INoqw==\\n=ua5l\\n-----END PGP MESSAGE-----\\n")
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	fmt.Println("解密key：", key)
	//	msg, err := en.DecryptMessageWithKey("-----BEGIN PGP MESSAGE-----\\n\\nwy4ECQMIoL/bAHnpmd/gZP6lR1LTZGHUx3iYSKK1rxB1J1UzM+PtVtnd0lNJ\\neozv0jQBPzE9IBiwCNOhPlFfGCpj60aDrDyCHq49m7BPpggZ7E0aEFKKd64a\\nuCGBLBZ2knRmp6we\\n=ddUS\\n-----END PGP MESSAGE-----\\n", key)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	fmt.Println("解密后消息：", msg)
	//}
	//key, err := en.DecryptMessageWithKey("wy4ECQMIgjl580Ix0lzgp+pM4vvr+tdrDKHGsg2ofmissf9ieX2ExcdBMUBtjf+q\n0l4B0OpkdfBevPKTlA2kSYHbP8zOuXiQXBph1DRwZfWSt0zv/LiPBdehBtnuSe8P\n9VRxhVn18EB4GkQFSRvZmk0xMp5pFz+Ei00f7JNJnK1ciTP1W9zroZ/ir5qLf0+N\n=zsDa", "123")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(key)
}
