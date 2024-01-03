package table

// InitDatabase 初始化数据库结构体
type InitDatabase struct {
	ID       string
	Migrate  any
	Rollback any
}
