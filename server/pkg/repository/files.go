package repository

import (
	"database/sql"
	"server/pkg/constant"
	"server/pkg/model"
)

type FilesRepository struct {
	DB *sql.DB
}

func NewFilesRepository(DB *sql.DB) *FilesRepository {
	return &FilesRepository{DB: DB}
}

func (f *FilesRepository) GetFiles() (*[]model.File, error) {
	query := "SELECT COUNT(*) from file"
	row := f.DB.QueryRow(query)
	var count int64

	if err := row.Scan(&count); err != nil {
		return nil, err
	}

	files := make([]model.File, count)
	query = "SELECT * FROM file;"

	rows, err := f.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		if err := rows.Scan(&files[i].Id,
			&files[i].Uid, &files[i].PathToFile, &files[i].Meta); err != nil {
			return nil, err
		}
		i++
	}
	return &files, nil
}

func (f *FilesRepository) GetFile(uid model.Uid) (*model.File, error) {
	query := "SELECT * FROM file WHERE uid = $1;"
	row := f.DB.QueryRow(query, uid)
	file := new(model.File)

	if err := row.Scan(&file.Id, &file.Uid, &file.PathToFile, &file.ChunkSize, &file.Meta); err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FilesRepository) AddFile(file *model.File) (model.Uid, error) {
	query := "INSERT INTO file (uid, path_to_file, chunk_size, meta) VALUES ($1, $2, $3, $4) RETURNING uid"
	row := f.DB.QueryRow(query, file.Uid, file.PathToFile, file.ChunkSize, file.Meta)
	var uid model.Uid = 0

	if err := row.Scan(&uid); err != nil {
		return constant.ErrUID, err
	}
	return uid, nil
}

func (f *FilesRepository) CheckUID(uid model.Uid) (bool, error) {
	query := "SELECT COUNT(*) FROM file WHERE uid = $1"
	row := f.DB.QueryRow(query, uid)
	var isExists = -1

	if err := row.Scan(&isExists); err != nil {
		return true, err
	}

	if isExists == 1 {
		return true, nil
	}
	return false, nil
}
