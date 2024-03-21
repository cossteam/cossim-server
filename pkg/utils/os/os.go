package os

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetPackagePath() (string, error) {
	appPath, err := os.Getwd()

	if err != nil {
		return "", err

	}
	lastIndex := strings.LastIndex(appPath, "/coss-server/")
	if lastIndex == -1 {
		return "", fmt.Errorf("无法获取到coss-server的路径")
	} else {
		matchingPath := appPath[0 : lastIndex+len("/coss-server/")]
		return matchingPath, nil
	}
}

// 获取出口IP
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
