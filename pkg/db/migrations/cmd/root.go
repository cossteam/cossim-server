package cmd

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/db/migrations/config"
	"github.com/cossim/coss-server/pkg/db/migrations/core"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "database migration",
	Long:  `UserShow how to use database migration`,
	PreRun: func(cmd *cobra.Command, args []string) {
		//运行
		fmt.Println("加载数据库")
		if err := config.Init(); err != nil {
			panic(err)
		}
		if err := core.InitDB(&config.C); err != nil {
			fmt.Println("数据库加载失败", err)
			os.Exit(1)
		}
		fmt.Println("初始化数据库表")
		if err := core.Init(); err != nil {
			fmt.Println("数据库初始化失败", err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("迁移完成")

	},
}

// 开始运行
func Execute() {
	// 加载环境变量
	// 调用Load函数来加载.env文件

	// 运行
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
