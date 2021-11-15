package repository

import (
	"database/sql"
	"server/pkg/model"
)

type Files interface {
	GetFiles() (*[]model.File, error)
	GetFile(uid model.Uid) (*model.File, error)
	AddFile(f *model.File) (model.Uid, error)
	CheckUID(uid model.Uid) (bool, error)
}

type Repository struct {
	Files
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{Files: NewFilesRepository(db)}
}
