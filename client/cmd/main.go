package main

import (
	"client"
	"client/pkg/handler"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	Client := new(client.Client)
	err := Client.Run("tcp", ServerAddr, new(handler.Handler))
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if err = Client.Shutdown(); err != nil {
		log.Fatalf("%s", err.Error())
	}
}

var ServerAddr string

func init() {
	err := initEnv()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	ServerAddr = os.Getenv("SERVER_ADDR") + ":" + os.Getenv("SERVER_PORT")
}

func initEnv() error{
	return godotenv.Load(".env")
}
