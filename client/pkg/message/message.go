package message

import (
	"client/pkg/model"
	"time"
)

type Message struct {
	Command string
	Uid model.Uid
	FileInfo struct {
		Name string
		Size int64
		ModTime time.Time
		ChunkSize int
	}
	Err string
}