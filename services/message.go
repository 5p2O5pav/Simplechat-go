package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"
	"chat-system-go/internal/config"
	"chat-system-go/internal/models"
	"chat-system-go/internal/repository"
	"chat-system-go/internal/utils"
)

func SaveUserMessage(sid, msgText string, fileData io.Reader, fileName, mimeType string, fileSize int64) (*models.Message, error) {
	now := time.Now().UnixMilli()
	msgID := generateID()
	msg := &models.Message{
		ID:    msgID,
		Sid:   sid,
		Role:  "user",
		Time:  now,
		Type:  "text",
	}
	if fileData != nil {
		// save file
		ext := filepath.Ext(fileName)
		if ext == "" {
			ext = ".bin"
		}
		saveName := utils.GenerateUniqueFileName(ext)
		destPath := filepath.Join("public/uploads", saveName)
		if _, err := utils.SaveUploadedFile(fileData, destPath); err != nil {
			return nil, err
		}
		fileURL := "/uploads/" + saveName
		if config.AppConfig.Domain != "" {
			fileURL = "https://" + config.AppConfig.Domain + fileURL
		}
		// determine type
		msg.Type = utils.GetFileCategory(mimeType)
		msg.FileName = fileName
		msg.FileSize = fileSize
		msg.FileURL = fileURL
		msg.MimeType = mimeType
		if msgText != "" {
			msg.Text = msgText
		}
		// record file in files table
		repository.InsertFileRecord(sid, destPath, fileSize)
	} else {
		msg.Text = msgText
	}
	if err := repository.InsertMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func SaveAgentMessage(sid, text, fileURL, fileName, mimeType, msgType string) (*models.Message, error) {
	now := time.Now().UnixMilli()
	msgID := generateID()
	msg := &models.Message{
		ID:       msgID,
		Sid:      sid,
		Role:     "agent",
		Time:     now,
		Type:     msgType,
		Text:     text,
		FileURL:  fileURL,
		FileName: fileName,
		MimeType: mimeType,
	}
	if err := repository.InsertMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
