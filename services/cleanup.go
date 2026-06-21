package services

import (
	"log"
	"time"
	"chat-system-go/internal/repository"
	"chat-system-go/internal/utils"
)

func CleanupExpiredSessions() error {
	now := time.Now().UnixMilli()
	sids, err := repository.ListExpiredSessions(now)
	if err != nil {
		return err
	}
	for _, sid := range sids {
		// 获取该会话的所有文件
		files, err := repository.GetFilesBySid(sid)
		if err != nil {
			log.Printf("Error getting files for sid %s: %v", sid, err)
			continue
		}
		for _, f := range files {
			if err := utils.DeleteFile(f.Path); err != nil {
				log.Printf("Delete file %s error: %v", f.Path, err)
			}
		}
		// 删除数据库记录（级联删除会自动删除 messages, files, topics, device_names, autoreply_flags）
		if err := repository.DeleteSession(sid); err != nil {
			log.Printf("Delete session %s error: %v", sid, err)
		}
		// 可选：删除 Telegram 话题（单独 API 调用，非必须）
	}
	// 清理过期的 blocked_ips
	if err := repository.DeleteExpiredBlockedIPs(now); err != nil {
		log.Printf("Cleanup blocked IPs error: %v", err)
	}
	return nil
}
