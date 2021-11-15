package handler

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"os"
	"server/logs"
	"server/pkg/constant"
	"server/pkg/message"
	"server/pkg/service"
)

const (
	UPLOAD="upload"
	UNLOAD="unload"
)

type FilesHandler struct {
	Service *service.Service
}

func (h *Handler) addFile(connectionIn *bufio.Reader, connectionOut *bufio.Writer, requestMessage *message.Message) {
	file, uid, err := h.Service.AddFile(requestMessage)
	if err != nil {
		log.Printf("%s\n", err.Error())
	}
	defer file.Close()

	if err := receiveFile(connectionIn, file); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	responseMessage := new(message.Message)
	responseMessage.Uid = uid
	if err := sendMessageInfo(connectionOut, responseMessage); err != nil {
		log.Printf("%s", err.Error())
		return
	}
}

func (h *Handler) getFile(connectionOut *bufio.Writer, requestMessage *message.Message) {
	uid := requestMessage.Uid
	responseMessage := new(message.Message)

	file, err := h.Service.GetFile(uid)
	if err != nil {
		log.Printf("%s", err.Error())
		responseMessage.Err = "uid was not found"
		if err := sendMessageInfo(connectionOut, responseMessage); err != nil {
			log.Printf("%s\n", err.Error())
			return
		}
	}

	responseMessage.Uid = uid
	responseMessage.FileInfo.ChunkSize = file.ChunkSize
	if err := fillMessage(responseMessage, file.PathToFile); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	log.Println(logs.LogFormat{
		Uid:      responseMessage.Uid,
		FileName: responseMessage.FileInfo.Name,
		Size:     responseMessage.FileInfo.Size,
		Mode:     requestMessage.Command,
	})

	if err := sendMessageInfo(connectionOut, responseMessage); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	if err := sendFile(connectionOut, file.PathToFile, responseMessage.FileInfo.ChunkSize); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}
}

func sendFile(connectionOut *bufio.Writer, pathToFile string, chunkSize int) error {
	file, err := os.Open(pathToFile)
	if err != nil {
		return err
	}
	defer file.Close()

	buff := make([]byte, chunkSize)
	for {
		countBytes, err := file.Read(buff)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var pBuff *[]byte
		if countBytes == chunkSize {
			pBuff = &buff
		} else
		if countBytes != chunkSize {
			incompleteBuff := make([]byte, countBytes)
			copy(incompleteBuff, buff)
			pBuff = &incompleteBuff
		}

		if _, err = connectionOut.Write(*pBuff); err != nil {
			return err
		}

		if err = connectionOut.Flush(); err != nil {
			return err
		}
	}

	if _, err = connectionOut.Write([]byte("EOF")); err != nil {
		return err
	}

	if err = connectionOut.Flush(); err != nil {
		return err
	}

	return nil
}

func receiveFile(connectionIn *bufio.Reader, file *os.File) error {
	buff := make([]byte, constant.ChuckSize)
	var pBuff *[]byte
	for {
		countBytes, err := connectionIn.Read(buff)
		if err == io.EOF {
			break
		} else
		if err != nil {
			return err
		}

		if countBytes == constant.ChuckSize {
			pBuff = &buff
		} else
		if countBytes != constant.ChuckSize {
			incompleteBuff := make([]byte, countBytes)
			copy(incompleteBuff, buff)
			pBuff = &incompleteBuff
		}

		if bytes.Equal((*pBuff)[len(*pBuff)- len(constant.EOF):], []byte(constant.EOF)) {
			if len(*pBuff) != 3 {
				if _, err = file.Write((*pBuff)[:len(*pBuff)- len(constant.EOF)]); err != nil {
					return err
				}
			}
			break
		}

		if _, err = file.Write(*pBuff);  err != nil {
			return err
		}
	}

	return nil
}

func fillMessage(message *message.Message, pathToFile string) error {
	fileInfo, err := os.Stat(pathToFile)
	if err != nil {
		return err
	}

	message.FileInfo.Name = fileInfo.Name()
	message.FileInfo.Size = fileInfo.Size()
	message.FileInfo.ModTime = fileInfo.ModTime()

	return nil
}

func sendMessageInfo(connectionOut *bufio.Writer, message *message.Message) error {
	var networkOut bytes.Buffer
	encoder := gob.NewEncoder(&networkOut)

	if err := encoder.Encode(&message); err != nil {
		return err
	}

	if _, err := connectionOut.Write(networkOut.Bytes()); err != nil {
		return err
	}

	if err := connectionOut.Flush(); err != nil {
		return err
	}

	return nil
}
