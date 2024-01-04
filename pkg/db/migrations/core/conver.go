package core

import (
	"github.com/cossim/coss-server/pkg/db/migrations/table"
	"github.com/go-gormigrate/gormigrate/v2"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// convertToKebabCase 将驼峰名转成-分隔
func convertToKebabCase(str string) string {
	var result strings.Builder
	//遍历字符串
	for i, ch := range str {
		//如果遇到大写
		if unicode.IsUpper(ch) {
			//忽略第一个单词
			if i > 0 {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(ch))
		} else {
			//处理小写
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func Init() error {
	//反射值 可以调用
	value := reflect.ValueOf(table.InitDatabase{})
	//反射类型 拿到定义的东西
	structType := reflect.TypeOf(table.InitDatabase{})

	//存储所有表
	defaultMigration := []*gormigrate.Migration{}
	//遍历结构体所有方法
	for i := 0; i < structType.NumMethod(); i++ {
		method := structType.Method(i)
		methodValue := value.MethodByName(method.Name)
		result := methodValue.Call(nil) //调用方法拿到迁移对象
		//类型断言转换成对应类型
		resultValue := result[0].Interface().(*gormigrate.Migration)
		if resultValue.ID == "" {
			resultValue.ID = convertToKebabCase(method.Name)
		} else {
			resultValue.ID = resultValue.ID + "-" + convertToKebabCase(method.Name)
		}
		//把表添加到列表
		defaultMigration = append(defaultMigration, resultValue)
	}
	sort.Slice(defaultMigration, func(i, j int) bool {
		//根据字符串大小排序，如果数字部分一样，则按照字母大小
		return defaultMigration[i].ID < defaultMigration[j].ID
	})
	// 创建迁移实例
	migrator := gormigrate.New(DB, gormigrate.DefaultOptions, defaultMigration)
	if err := migrator.Migrate(); err != nil {
		return err
	}
	return nil
}
