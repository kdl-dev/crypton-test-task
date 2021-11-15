package service

import (
	"math/rand"
	"os"
	"server/pkg/message"
	"server/pkg/model"
	"server/pkg/repository"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())	// for generate UID
}

type Files interface {
	GetFiles() (*[]model.File, error)
	GetFile(uid model.Uid) (*model.File, error)
	AddFile(message *message.Message) ( *os.File,model.Uid, error)
	GenerateUID() (model.Uid, error)
}

type Service struct {
	Files
}

func NewService(repository *repository.Repository) *Service {
	return &Service{Files: NewFilesService(repository)}
}
