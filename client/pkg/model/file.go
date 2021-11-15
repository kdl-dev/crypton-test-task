package model

type Uid int32

type File struct {
	Id Uid
	Uid Uid
	PathToFile string
	ChunkSize int
	Meta string
}
