package services

import (
	"time"
	"chat-system-go/internal/config"
	"chat-system-go/internal/models"
	"chat-system-go/internal/repository"
)

func IsBlocked(ip string) (bool, error) {
	b, err := repository.GetBlockedIP(ip)
	if err != nil {
		return false, err
	}
	if b == nil {
		return false, nil
	}
	if b.ExpireAt < time.Now().UnixMilli() {
		return false, nil
	}
	return true, nil
}

func CheckRateLimit(ip string) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - (now % 60000) // minute aligned

	rl, err := repository.GetRateLimit(ip, windowStart)
	if err != nil {
		return false, err
	}
	if rl == nil {
		rl = &models.RateLimit{
			IP:          ip,
			WindowStart: windowStart,
			Count:       1,
			UpdatedAt:   now,
		}
		if err := repository.UpsertRateLimit(rl); err != nil {
			return false, err
		}
		return false, nil
	}
	if rl.Count >= config.AppConfig.RateLimitCount {
		// block
		banHours := config.AppConfig.RateLimitBanHours
		expireAt := now + int64(banHours*3600*1000)
		if err := repository.UpsertBlockedIP(ip, now, expireAt); err != nil {
			return false, err
		}
		return true, nil
	}
	// increment
	rl.Count++
	rl.UpdatedAt = now
	if err := repository.UpsertRateLimit(rl); err != nil {
		return false, err
	}
	return false, nil
}
