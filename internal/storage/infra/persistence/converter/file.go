package converter

import (
	"github.com/cossim/coss-server/internal/storage/domain/entity"
	"github.com/cossim/coss-server/internal/storage/infra/persistence/po"
)

func FileEntityToPO(e *entity.File) *po.File {
	return &po.File{
		ID:        e.ID,
		Owner:     e.Owner,
		Name:      e.Name,
		Size:      e.Size,
		Type:      uint(e.Type),
		Status:    uint(e.Status),
		Provider:  string(e.Provider),
		Content:   e.Content,
		Path:      e.Path,
		Share:     e.Share,
		CreatedAt: e.CreatedAt,
	}
}

func FilePOToEntity(po *po.File) *entity.File {
	return &entity.File{
		ID:        po.ID,
		Owner:     po.Owner,
		Name:      po.Name,
		Size:      po.Size,
		Type:      entity.FileType(po.Type),
		Status:    entity.FileStatus(po.Status),
		Provider:  entity.Provider(po.Provider),
		Content:   po.Content,
		Path:      po.Path,
		Share:     po.Share,
		CreatedAt: po.CreatedAt,
	}
}
