package handler

import (
	"bufio"
	"log"
	"os"
)

const (
	UPLOAD="upload"
	UNLOAD="unload"
)

type Handler struct {}

func (h *Handler) Handle(connectionIn *bufio.Reader, connectionOut *bufio.Writer) {
	if len(os.Args) == 1 {
		log.Fatal("3 arguments were expected.")
	}

	switch os.Args[1] {
	case UPLOAD:
		if len(os.Args) < 4 {
			log.Fatal("3 arguments were expected.")
		}
		if err := h.Upload(connectionIn, connectionOut); err != nil {
			log.Fatalf("%s", err.Error())
		}
	case UNLOAD:
		if len(os.Args) < 5 {
			log.Fatal("4 arguments were expected.")
		}
		if err := h.Unload(connectionIn, connectionOut); err != nil {
			log.Fatalf("%s", err.Error())
		}
	default:
		log.Fatal("Unsupported command")
	}
}
