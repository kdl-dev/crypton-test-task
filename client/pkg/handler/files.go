package handler

import (
	"bufio"
	"bytes"
	"client/pkg/constant"
	"client/pkg/encryption"
	"client/pkg/message"
	"client/pkg/model"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

func (h *Handler) Upload(connectionIn *bufio.Reader, connectionOut *bufio.Writer) error {
	pathToFile := os.Args[2]
	password := os.Args[3]

	if err := validateFile(pathToFile); err != nil {
		return err
	}

	fileInfo, _ := os.Stat(pathToFile)
	chunkSize := int(math.Min(constant.ChuckSize, float64(fileInfo.Size())))

	requestMessage := new(message.Message)
	requestMessage.Command = UPLOAD
	requestMessage.FileInfo.ChunkSize = encryption.EncodeChunkSize(chunkSize, password)

	if err := fillMessage(requestMessage, pathToFile); err != nil {
		return err
	}

	if err := sendMessageInfo(connectionOut, requestMessage); err != nil {
		return err
	}

	if err := sendFile(connectionOut, pathToFile, password); err != nil {
		return err
	}

	responseMessage, err := receiveMessageInfo(connectionIn)
	if err != nil {
		return err
	}

	log.Printf("UID = %d\n", responseMessage.Uid)
	return nil
}

func (h *Handler) Unload(connectionIn *bufio.Reader, connectionOut *bufio.Writer) error {
	uid := os.Args[2]
	pathToDir := os.Args[3]
	password := os.Args[4]

	if err := validateUid(uid); err != nil {
		return err
	}

	if err := validateDir(pathToDir); err != nil {
		return err
	}

	Uid, _ := strconv.Atoi(uid)
	requestMessage := new(message.Message)
	requestMessage.Command = UNLOAD
	requestMessage.Uid = model.Uid(Uid)

	if err := sendMessageInfo(connectionOut, requestMessage); err != nil {
		return err
	}

	responseMessage, err := receiveMessageInfo(connectionIn)
	if err != nil {
		return err
	}

	if responseMessage.Err != "" {
		return errors.New(responseMessage.Err)
	}

	file, err := createFile(pathToDir, responseMessage.FileInfo.Name)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = receiveFile(file, connectionIn, responseMessage.FileInfo.ChunkSize, password); err != nil {
		return err
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
	var networkOutBuff bytes.Buffer
	encoder := gob.NewEncoder(&networkOutBuff)

	if err := encoder.Encode(&message); err != nil {
		return err
	}

	if _, err := connectionOut.Write(networkOutBuff.Bytes()); err != nil {
		return err
	}

	if err := connectionOut.Flush(); err != nil {
		return err
	}

	return nil
}

func receiveMessageInfo(connectionIn *bufio.Reader) (*message.Message, error) {
	message := new(message.Message)
	decoder := gob.NewDecoder(connectionIn)
	if err := decoder.Decode(&message); err != nil {
		return nil, err
	}

	return message, nil
}

func sendFile(connectionOut *bufio.Writer, pathToFile string, password string) error {
	file, err := os.Open(pathToFile)
	if err != nil {
		return err
	}
	defer file.Close()

	buff := make([]byte, constant.ChuckSize)
	for {
		countBytes, err := file.Read(buff)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var pBuff *[]byte
		if countBytes == constant.ChuckSize {
			pBuff = &buff
		} else
		if countBytes != constant.ChuckSize {
			incompleteBuff := make([]byte, countBytes)
			copy(incompleteBuff, buff)
			pBuff = &incompleteBuff
		}

		encryptedBuff, err := encryption.Encode(*pBuff, password, encryption.Cbc)
		if err != nil {
			return err
		}

		if _, err = connectionOut.Write(encryptedBuff); err != nil {
			return err
		}

		if err = connectionOut.Flush(); err != nil {
			return err
		}
	}

	if _, err = connectionOut.Write([]byte(constant.EOF)); err != nil {
		return err
	}

	if err = connectionOut.Flush(); err != nil {
		return err
	}

	return nil
}

func receiveFile(file *os.File, in *bufio.Reader, encodeChunkSize int, password string) error {
	buff := make([]byte, encodeChunkSize)
	for {
		countBytes, err := io.ReadFull(in, buff)
		if err != nil {
			if err == io.EOF {
				break
			}
			if countBytes == 0 {
				return err
			}
		}

		var pBuff *[]byte
		if countBytes == encodeChunkSize {
			pBuff = &buff
		} else
		if countBytes != encodeChunkSize {
			if bytes.Equal(buff, []byte(constant.EOF)) {
				break
			}

			incompleteBuff := make([]byte, countBytes)
			copy(incompleteBuff, buff)
			pBuff = &incompleteBuff
		}

		if bytes.Equal((*pBuff)[len(*pBuff)- len(constant.EOF):], []byte(constant.EOF)) {
			if len(*pBuff) != 3 {
				decryptedBuff, err := encryption.Decode((*pBuff)[:len(*pBuff)- len(constant.EOF)], password, encryption.Cbc)
				if err != nil {
					return err
				}

				if _, err = file.Write(decryptedBuff); err != nil {
					return err
				}
			}
			break
		}

		decryptedBuff, err := encryption.Decode(*pBuff, password, encryption.Cbc)
		if err != nil {
			return err
		}

		if _, err = file.Write(decryptedBuff); err != nil {
			return err
		}
	}
	return nil
}

func validateFile(pathToFile string) error {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return errors.New("file is not exist")
	}
	return nil
}

func validateDir(pathToDir string) error {
	fileInfo, err := os.Stat(pathToDir)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return errors.New(pathToDir + " is not directory")
	}
	return nil
}

func validateUid(uid string) error {
	Uid, err := strconv.Atoi(uid)
	if err != nil {
		return err
	}

	if Uid < 1 {
		return errors.New("uid must be greater than zero")
	}

	return nil
}

func createFile(pathToStorage, fileName string) (*os.File, error){
	var filePath string
	index := 1

	if pathToStorage[len(pathToStorage) - 1] != filepath.Separator {
		pathToStorage += string(filepath.Separator)
	}

	for {
		filePath = fmt.Sprintf("%s%s",
			pathToStorage , fileName)
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

func fileExists(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !fileInfo.IsDir()
}
