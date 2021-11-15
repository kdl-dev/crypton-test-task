package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"server"
	"server/pkg/handler"
	"server/pkg/repository"
	"server/pkg/service"
	"strconv"
	"syscall"
)

func main() {
	Server := new(server.Server)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := Server.Run("tcp", ":" + Port, Handler, ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	}()

	sigCh := make(chan os.Signal)
	defer close(sigCh)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sigCh

	if err := Server.Shutdown(cancel); err != nil {
		log.Printf("error occured on server shutting down: %s\n", err.Error())
	}

	if err := db.Close(); err != nil {
		log.Printf("error occured on db connection close: %s", err.Error())
	}
}

var Port string
var db *sql.DB
var repos *repository.Repository
var services *service.Service
var Handler *handler.Handler

func init() {
	var err error
	if err = initEnv(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	countOfArgs := len(os.Args)

	if countOfArgs == 1 {
		Port = os.Getenv("SERVER_PORT")
	} else
	if countOfArgs > 1 {
		Port = os.Args[1]
	}

	if err = ValidatePort(Port); err != nil {
		log.Fatalf("%s", err.Error())
	}

	initLogs()
	initDependency()
}

func initEnv() error{
	return godotenv.Load(".env")
}

func initLogs() {
	logPath := os.Getenv("LOG_PATH")
	var fileLog *os.File

	if 	err := os.MkdirAll(os.Getenv("STORAGE_ROOT_PATH"), 0750); err != nil {
		log.Fatalf("%s", err.Error())
	}

	_, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		fileLog, err = os.Create(logPath)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	} else {
		fileLog, err = os.OpenFile(logPath,  os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	}
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	log.SetOutput(fileLog)
}

func initDependency() {
	var err error
	db, err = repository.NewPostgresDB(repository.Config{
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"),
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err.Error())
	}

	repos = repository.NewRepository(db)
	services = service.NewService(repos)
	Handler = handler.NewHandler(services)
}

func ValidatePort(port string) error {
	validPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	if validPort < 1024 && validPort > 65535 {
		return errors.New("")
	}

	return nil
}
