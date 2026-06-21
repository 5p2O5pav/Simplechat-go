package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
	"chat-system-go/internal/config"
	"chat-system-go/internal/models"
	"chat-system-go/internal/repository"
)

func NewSession() (sid, token string, err error) {
	sidBytes := make([]byte, 16)
	if _, err := rand.Read(sidBytes); err != nil {
		return "", "", err
	}
	sid = hex.EncodeToString(sidBytes)
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(tokenBytes)
	err = repository.CreateSession(sid, token, config.AppConfig.ExpireDays)
	return
}

func ValidateSession(sid, token string) (bool, error) {
	s, err := repository.GetSession(sid)
	if err != nil {
		return false, err
	}
	if s == nil {
		return false, nil
	}
	if s.Token != token {
		return false, nil
	}
	if s.ExpireAt < time.Now().UnixMilli() {
		return false, nil
	}
	// update last active
	repository.UpdateSessionLastActive(sid)
	return true, nil
}

func GetSessionBySid(sid string) (*models.Session, error) {
	return repository.GetSession(sid)
}

func RefreshSessionExpiry(sid string) error {
	expire := time.Now().UnixMilli() + int64(config.AppConfig.ExpireDays*24*3600*1000)
	_, err := repository.DB.Exec(`UPDATE sessions SET expire_at = ? WHERE sid = ?`, expire, sid)
	return err
}
