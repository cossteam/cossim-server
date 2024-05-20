package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdatesUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUser(ctx context.Context, userID string) (*entity.User, error)
	GetWithOptions(ctx context.Context, query *entity.User) (*entity.User, error)
	ListUser(ctx context.Context, query *entity.ListUserOptions) ([]*entity.User, error)
	InsertAndUpdateUser(ctx context.Context, e *entity.User) error
}
