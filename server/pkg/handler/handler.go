package handler

import (
	"bufio"
	"encoding/gob"
	"log"
	"server/logs"
	"server/pkg/message"
	"server/pkg/service"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{Service : service}
}

func (h *Handler) Handle(connectionIn *bufio.Reader, connectionOut *bufio.Writer) {
	requestMessage, err := receiveMessageInfo(connectionIn)
	if err != nil {
		return
	}

	switch requestMessage.Command {
	case UPLOAD:
		log.Println(logs.LogFormat{
			Uid:      -1,
			FileName: requestMessage.FileInfo.Name,
			Size:     requestMessage.FileInfo.Size,
			Mode:     requestMessage.Command,
		})
		h.addFile(connectionIn, connectionOut, requestMessage)
	case UNLOAD:
		h.getFile(connectionOut, requestMessage)
	default:
		log.Println("unsupported command")
	}
}

func receiveMessageInfo(connectionIn *bufio.Reader) (*message.Message, error) {
	message := new(message.Message)
	decoder := gob.NewDecoder(connectionIn)
	if err := decoder.Decode(&message); err != nil {
		return nil, err
	}

	return message, nil
}
