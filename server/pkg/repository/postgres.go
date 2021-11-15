package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	File = "file"
)

type Config struct {
	Host string
	Port string
	Username string
	Password string
	DBName string
}

func NewPostgresDB(cfg Config) (*sql.DB, error){
	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName,cfg.Password))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
