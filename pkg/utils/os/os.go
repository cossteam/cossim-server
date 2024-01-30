package os

import (
	"fmt"
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
