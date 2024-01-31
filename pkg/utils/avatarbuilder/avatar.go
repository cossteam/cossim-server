package avatarbuilder

import (
	"github.com/shiningrush/avatarbuilder"
	"github.com/shiningrush/avatarbuilder/calc"
	"image/color"
	"math/rand"
	"unicode"
)

var colors = []uint32{
	0xff6200, 0x42c58e, 0x5a8de1, 0x785fe0,
	0xf6546a, 0x2ecc71, 0x3498db, 0xffdb4d,
	0x9b59b6, 0xe74c3c, 0x1abc9c, 0xf39c12,
	// 添加更多的颜色值...
}

func GenerateAvatar(name string, path string) ([]byte, error) {
	ab := avatarbuilder.NewAvatarBuilder(path+"SourceHanSansSC-Medium.ttf", &calc.SourceHansSansSCMedium{})
	ab.SetBackgroundColorHex(GetRandomColor())
	ab.SetFrontgroundColor(color.White)
	ab.SetFontSize(80)
	ab.SetAvatarSize(200, 200)
	image, err := ab.GenerateImage(GetInitials(name))
	if err != nil {
		return nil, err
	}
	return image, nil
}

// 随机返回颜色
func GetRandomColor() uint32 {
	// 生成一个随机数，作为颜色索引
	index := rand.Intn(len(colors))
	return colors[index]
}

// 判断是否中文
func IsName(str string) bool {
	// 所有字符都是中文字符
	for _, char := range str {
		if !unicode.Is(unicode.Scripts["Han"], char) {
			return false
		}
	}

	// 长度至少为2
	if len(str) < 2 {
		return false
	}

	return true
}

// 获取首字母或者姓名的后两个中文字符
func GetInitials(name string) string {
	var initials string
	runes := []rune(name)
	if len(runes) < 2 {
		return name
	}
	if IsName(name) {
		// 取后两个字符
		initials = string(runes[len(runes)-2:])
	} else {
		if len(runes) < 2 {
			initials = string(runes[0])
		} else {
			initials = string(runes[0:2])
		}
	}

	return initials
}
