package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"chat-system-go/internal/config"
	"chat-system-go/internal/services"
	"chat-system-go/internal/utils"
)

func SendHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(config.AppConfig.MaxFileSize + 1024); err != nil {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}
	sid := r.FormValue("sid")
	token := r.FormValue("token")
	msgText := r.FormValue("msg")

	if sid == "" || token == "" {
		http.Error(w, "Missing sid or token", http.StatusBadRequest)
		return
	}
	valid, err := services.ValidateSession(sid, token)
	if err != nil || !valid {
		http.Error(w, "Invalid session", http.StatusForbidden)
		return
	}

	ip := utils.GetRealIP(r)
	blocked, err := services.IsBlocked(ip)
	if err == nil && blocked {
		// 静默成功
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "msgId": "blocked", "msgData": struct{}{}})
		return
	}
	rateLimited, err := services.CheckRateLimit(ip)
	if err == nil && rateLimited {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "msgId": "limited", "msgData": struct{}{}})
		return
	}

	// 获取话题
	topicID, err := services.GetOrCreateTopic(sid, r.UserAgent())
	if err != nil {
		http.Error(w, "Failed to create topic", http.StatusInternalServerError)
		return
	}

	// 处理文件
	var fileReader io.Reader
	var fileName, mimeType string
	var fileSize int64
	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		fileSize = header.Size
		fileName = header.Filename
		mimeType = header.Header.Get("Content-Type")
		fileReader = file
	}

	// 保存消息
	msg, err := services.SaveUserMessage(sid, msgText, fileReader, fileName, mimeType, fileSize)
	if err != nil {
		http.Error(w, "Save message error", http.StatusInternalServerError)
		return
	}

	// 更新最后活动
	services.UpdateSessionLastActive(sid)

	// 异步发送到 Telegram
	go forwardToTelegram(topicID, msg, ip)

	// 触发自动回复 (延迟 20s)
	go scheduleAutoReply(sid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"msgId":   msg.ID,
		"msgData": msg,
	})
}

func forwardToTelegram(topicID int64, msg *models.Message, ip string) {
	// 构造文本
	var text string
	if msg.Type == "text" {
		text = msg.Text
	} else {
		text = fmt.Sprintf("📎 文件: %s", msg.FileURL)
		if msg.Text != "" {
			text = msg.Text + "\n\n" + text
		}
	}
	// 添加 IP 信息
	geo := utils.GetGeoInfo(ip) // 需实现，可省略
	text += fmt.Sprintf("\n\n[IP: %s | %s]", ip, geo)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.AppConfig.BotToken)
	payload := map[string]interface{}{
		"chat_id":            config.AppConfig.ChatID,
		"message_thread_id":  topicID,
		"text":               text,
	}
	jsonData, _ := json.Marshal(payload)
	http.Post(url, "application/json", strings.NewReader(string(jsonData)))
}

func scheduleAutoReply(sid string) {
	time.Sleep(20 * time.Second)
	ok, err := services.ShouldSendAutoReply(sid)
	if err != nil {
		return
	}
	if ok {
		services.SendAutoReply(sid)
	}
}
