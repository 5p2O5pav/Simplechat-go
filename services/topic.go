package services

import (
	"encoding/json"
	"fmt"
	"time"
	"chat-system-go/internal/config"
	"chat-system-go/internal/models"
	"chat-system-go/internal/repository"
	"chat-system-go/internal/utils"
)

const maxSeq = 36 * 36 * 36 // 46656

func GenerateTopicName(deviceType string) (string, error) {
	seqKey := fmt.Sprintf("device:seq:%s", deviceType)
	seq, err := repository.IncrementGlobalSeq(seqKey)
	if err != nil {
		return "", err
	}
	seqNum := (seq - 1) % int64(maxSeq)
	for attempt := 0; attempt < 100; attempt++ {
		seqStr := utils.ToBase36(seqNum, 3)
		name := fmt.Sprintf("%s-%s", deviceType, seqStr)
		// 检查是否已被占用（直接查询 device_names 表中有无此 seq）
		// 注意：我们无法通过 device_names 直接查询 seq 是否存在，因为不同设备类型可重复。
		// 我们需要确保同一设备类型下的 seq 唯一。因此查询 device_names 中 device_type 和 seq 组合。
		var count int
		err := repository.DB.QueryRow(`SELECT COUNT(*) FROM device_names WHERE device_type = ? AND seq = ?`, deviceType, seqStr).Scan(&count)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return name, nil
		}
		seqNum = (seqNum + 1) % int64(maxSeq)
	}
	return "", fmt.Errorf("failed to allocate unique name for device %s", deviceType)
}

func CreateTopic(sid, deviceType string) (int64, string, error) {
	topicName, err := GenerateTopicName(deviceType)
	if err != nil {
		// fallback to old style
		topicName = fmt.Sprintf("用户 %s", sid[:8])
	}
	// 调用 Telegram API
	url := fmt.Sprintf("https://api.telegram.org/bot%s/createForumTopic", config.AppConfig.BotToken)
	payload := map[string]interface{}{
		"chat_id": config.AppConfig.ChatID,
		"name":    topicName,
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageThreadID int64 `json:"message_thread_id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", err
	}
	if !result.Ok {
		return 0, "", fmt.Errorf("telegram API error")
	}
	topicID := result.Result.MessageThreadID
	// 存储到数据库
	topic := &models.Topic{
		Sid:       sid,
		TopicID:   topicID,
		TopicName: topicName,
	}
	if err := repository.InsertTopic(topic); err != nil {
		return 0, "", err
	}
	// 存储设备名称
	if deviceType != "未知设备" {
		seqPart := topicName[len(deviceType)+1:]
		dev := &models.DeviceName{
			Sid:        sid,
			DeviceType: deviceType,
			Seq:        seqPart,
		}
		repository.InsertDeviceName(dev)
	}
	return topicID, topicName, nil
}

func GetOrCreateTopic(sid, ua string) (int64, error) {
	topic, err := repository.GetTopic(sid)
	if err != nil {
		return 0, err
	}
	if topic != nil {
		return topic.TopicID, nil
	}
	deviceType := utils.DetectDevice(ua)
	topicID, _, err := CreateTopic(sid, deviceType)
	return topicID, err
}
