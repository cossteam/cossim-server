package entity

type Admin struct {
	UserId    string
	Role      Role
	Status    AdminStatus
	ID        uint
	CreatedAt int64
	UpdatedAt int64
	DeletedAt int64
}

type Role uint

const (
	SuperAdminRole Role = 1 //超级管理员
	AdminRole      Role = 2 //管理员
)

type AdminStatus uint

const (
	NormalStatus   AdminStatus = 1
	DisabledStatus AdminStatus = 2
)

type Query struct {
	UserId *string
	Role   *Role
	Status *AdminStatus
	ID     *uint
}
