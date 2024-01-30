package encryption_test

import (
	"github.com/cossim/coss-server/pkg/utils/avatarbuilder"
	"github.com/cossim/coss-server/pkg/utils/os"
	"io/ioutil"
	"testing"
)

var colors = []uint32{
	0xff6200, 0x42c58e, 0x5a8de1, 0x785fe0,
}

func TestEncryption(t *testing.T) {
	// 导入你想要获取 init 方法所在路径的包
	//importedPackage := avatarbuilder.GenerateAvatar
	//
	//// 获取导入包的 reflect.Value
	//importedValue := reflect.ValueOf(importedPackage)
	//
	//// 获取导入包的 Type
	//importedType := importedValue.Type()
	//
	//fmt.Println("导入包的类型:", importedType)
	//// 获取导入包的路径
	//_ = importedType.PkgPath()
	//
	path, err := os.GetPackagePath()
	if err != nil {
		return
	}
	ava, err := avatarbuilder.GenerateAvatar("test", path)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("test.png", ava, 0644)
	if err != nil {
		t.Error(err)
	}

	//appPath := "//hitosea-005/Desktop/hitosea/coss-server/interface/relation"

}
