package persistence

//import (
//	"github.com/cossim/coss-server/internal/group/domain/group"
//	"gorm.io/gorm"
//)
//
//type Repositories struct {
//	Gr group.Repository
//	db *gorm.DB
//}
//
//func NewRepositories(db *gorm.DB) *Repositories {
//	return &Repositories{
//		Gr: NewGroupRepo(db),
//		db: db,
//	}
//}
//
//func (s *Repositories) Automigrate() error {
//	return s.db.AutoMigrate(&group.Group{})
//}
