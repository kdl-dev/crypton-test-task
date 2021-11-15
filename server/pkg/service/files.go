package service

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"server/pkg/constant"
	"server/pkg/message"
	"server/pkg/model"
	"server/pkg/repository"
	"sync"
)

type FilesService struct {
	repo *repository.Repository
}

func NewFilesService(repo *repository.Repository) *FilesService {
	return &FilesService{repo: repo}
}

func (f *FilesService) GetFiles() (*[]model.File, error) {
	return f.repo.Files.GetFiles()
}

func (f *FilesService) GetFile(uid model.Uid) (*model.File, error) {
	if uid < 1 {
		return nil, errors.New("uid must not be less than 1")
	}
	return f.repo.Files.GetFile(uid)
}

func (f *FilesService)AddFile(message *message.Message) (*os.File, model.Uid, error) {
	m := new(sync.Mutex)
	fileInfo :=  new(model.File)
	pathToStorage := os.Getenv("STORAGE_ROOT_PATH")
	fileName := message.FileInfo.Name

	m.Lock()
	file, err := createFile(pathToStorage, fileName)
	m.Unlock()
	if err != nil {
		return nil, constant.ErrUID, err
	}

	fileInfo, err = setFileInfo(fileInfo, file, message, f)
	if  err != nil {
		file.Close()

		if err = removeFile(fileInfo.PathToFile); err != nil {
			return nil, constant.ErrUID, err
		}

		return nil, constant.ErrUID, err
	}

	uid, err := f.repo.Files.AddFile(fileInfo)
	if  err != nil {
		file.Close()

		if err = removeFile(fileInfo.PathToFile); err != nil {
			return nil, 0, err
		}
		return nil, constant.ErrUID, err
	}
	return file, uid, nil
}

func (f *FilesService) GenerateUID() (model.Uid, error) {
	var uid model.Uid
	for {
		uid = model.Uid(1 + rand.Int31())
		isExists, err := f.repo.CheckUID(uid)
		if err != nil {
			return constant.ErrUID, err
		}
		if isExists {
			continue
		}
		break
	}
	return uid, nil
}

func createFile(pathToStorage, fileName string) (*os.File, error){
	var filePath string
	index := 1

	if pathToStorage[len(pathToStorage) - 1] != filepath.Separator {
		pathToStorage += string(filepath.Separator)
	}

	for {
		filePath = fmt.Sprintf("%s%s",
			pathToStorage, fileName)
		ext := filepath.Ext(filePath)

		if fileExists(filePath) {
			filePath = fmt.Sprintf("%s(%d)%s",
				filePath[:len(filePath) - len(ext)],
				index,
				filePath[len(filePath) - len(ext):])
			if fileExists(filePath) {
				index++
				continue
			}
		}
		break
	}
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func removeFile(pathToFile string) error {
	if fileExists(pathToFile) {
		return os.Remove(pathToFile)
	}

	return nil
}

func fileExists(pathToFile string) bool {
	fileInfo, err := os.Stat(pathToFile)
	if os.IsNotExist(err) {
		return false
	}
	return !fileInfo.IsDir()
}

func setFileInfo(file *model.File, out *os.File, message *message.Message, filesService *FilesService) (*model.File, error) {
	m := new(sync.Mutex)
	fi, err := out.Stat()
	if err != nil {
		return nil, err
	}

	file.PathToFile = out.Name()
	file.ChunkSize = message.FileInfo.ChunkSize
	file.Meta = fmt.Sprintf("Name: %s\nMode time: %s\nSize: %d MiB (%d KiB)\n",
		fi.Name(), message.FileInfo.ModTime, message.FileInfo.Size/1024/1024, message.FileInfo.Size/1024)

	m.Lock()
	file.Uid, err = filesService.GenerateUID()
	m.Unlock()

	if err != nil {
		return nil, err
	}

	return file, nil
}
