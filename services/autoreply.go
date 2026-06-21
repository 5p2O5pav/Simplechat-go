package services

import (
	"time"
	"chat-system-go/internal/config"
	"chat-system-go/internal/models"
	"chat-system-go/internal/repository"
)

func ShouldSendAutoReply(sid string) (bool, error) {
	// 检查是否已有 flag
	flag, err := repository.GetAutoReplyFlag(sid)
	if err != nil {
		return false, err
	}
	if flag != nil {
		// 如果 flag 在 1 小时内，则不再发送
		if time.Now().UnixMilli()-flag.SentAt < 3600*1000 {
			return false, nil
		}
	}
	// 检查最近 1 天内是否有客服消息
	count, err := repository.CountMessagesBySidAndRole(sid, "agent")
	if err != nil {
		return false, err
	}
	if count > 0 {
		// 检查最近一条客服消息是否在 24h 内
		// 简单：如果存在任何客服消息，且最近一条在 24h 内，则跳过
		// 但可能更准确：查询最近一条消息时间
		var lastAgentTime int64
		err := repository.DB.QueryRow(`SELECT time FROM messages WHERE sid = ? AND role = 'agent' ORDER BY time DESC LIMIT 1`, sid).Scan(&lastAgentTime)
		if err == nil && time.Now().UnixMilli()-lastAgentTime < 24*3600*1000 {
			return false, nil
		}
	}
	// 检查用户第一条消息是否在 30 分钟内
	firstUserTime, err := repository.GetLastUserMessageTime(sid) // 这是最早的用户消息
	if err != nil {
		return false, err
	}
	if firstUserTime == 0 {
		return false, nil
	}
	if time.Now().UnixMilli()-firstUserTime > 30*60*1000 {
		return false, nil
	}
	return true, nil
}

func SendAutoReply(sid string) error {
	now := time.Now().UnixMilli()
	// 插入两条预设消息
	messages := []string{
		"您好，当前是留言模式。客服暂时不在线，您的消息已成功传达，上线后会第一时间回复您，请稍候。",
		"如果有任何问题或需求可以先详细描述一下，方便客服上线后快速了解情况，帮您高效处理。",
	}
	for _, text := range messages {
		msg := &models.Message{
			ID:   generateID(),
			Sid:  sid,
			Role: "agent",
			Type: "text",
			Text: text,
			Time: time.Now().UnixMilli(),
		}
		if err := repository.InsertMessage(msg); err != nil {
			return err
		}
		time.Sleep(3 * time.Second) // 模拟原逻辑中的延时
	}
	// 设置 flag
	return repository.InsertAutoReplyFlag(sid, now)
}
