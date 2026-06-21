package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"chat-system-go/internal/config"
	"chat-system-go/internal/services"
	"chat-system-go/internal/repository"
	"chat-system-go/internal/utils"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	secret := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if secret != config.AppConfig.WebhookSecret {
		w.WriteHeader(http.StatusOK) // 静默返回
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	var update struct {
		Message *struct {
			MessageThreadID int64 `json:"message_thread_id"`
			From            struct {
				IsBot bool `json:"is_bot"`
			} `json:"from"`
			Text     string `json:"text"`
			Photo    []struct{ FileID string } `json:"photo"`
			Video    *struct{ FileID, FileName string } `json:"video"`
			Document *struct{ FileID, FileName, MimeType string } `json:"document"`
			Voice    *struct{ FileID, Duration int } `json:"voice"`
			Sticker  *struct{ FileID, Emoji string } `json:"sticker"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &update); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	msg := update.Message
	if msg == nil || msg.From.IsBot || msg.MessageThreadID == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	topicID := msg.MessageThreadID
	sid, err := repository.GetSidByTopicID(topicID)
	if err != nil || sid == "" {
		// fallback: scan topics
		w.WriteHeader(http.StatusOK)
		return
	}

	// 处理消息内容
	var text, fileURL, fileName, mimeType, msgType string
	if msg.Text != "" {
		text = msg.Text
		msgType = "text"
	}
	if len(msg.Photo) > 0 {
		fileID := msg.Photo[len(msg.Photo)-1].FileID
		fileURL, fileName, err = downloadTelegramFile(fileID, sid, ".jpg")
		if err == nil {
			msgType = "image"
			mimeType = "image/jpeg"
		}
	}
	if msg.Video != nil {
		fileID := msg.Video.FileID
		fileURL, fileName, err = downloadTelegramFile(fileID, sid, ".mp4")
		if err == nil {
			msgType = "video"
			mimeType = "video/mp4"
			fileName = msg.Video.FileName
		}
	}
	if msg.Document != nil {
		fileID := msg.Document.FileID
		ext := filepath.Ext(msg.Document.FileName)
		fileURL, fileName, err = downloadTelegramFile(fileID, sid, ext)
		if err == nil {
			msgType = "document"
			mimeType = msg.Document.MimeType
			fileName = msg.Document.FileName
		}
	}
	if msg.Voice != nil {
		fileID := strconv.Itoa(msg.Voice.FileID) // 实际是 string
		fileURL, fileName, err = downloadTelegramFile(fileID, sid, ".ogg")
		if err == nil {
			msgType = "voice"
		}
	}
	if msg.Sticker != nil {
		fileID := msg.Sticker.FileID
		fileURL, _, err = downloadTelegramFile(fileID, sid, ".webp")
		if err == nil {
			msgType = "sticker"
		}
	}

	// 保存客服消息
	agentMsg, err := services.SaveAgentMessage(sid, text, fileURL, fileName, mimeType, msgType)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	// 清除自动回复标志
	repository.DeleteAutoReplyFlag(sid)
	_ = agentMsg
	w.WriteHeader(http.StatusOK)
}

func downloadTelegramFile(fileID, sid, ext string) (fileURL, fileName string, err error) {
    // 1. 获取文件路径
    url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", config.AppConfig.BotToken, fileID)
    resp, err := http.Get(url)
    if err != nil {
        return "", "", err
    }
    defer resp.Body.Close()
    var result struct {
        Ok     bool `json:"ok"`
        Result struct {
            FilePath string `json:"file_path"`
        } `json:"result"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", "", err
    }
    if !result.Ok {
        return "", "", fmt.Errorf("telegram getFile error")
    }
    filePath := result.Result.FilePath
    // 2. 下载文件
    fileURLTg := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", config.AppConfig.BotToken, filePath)
    resp, err = http.Get(fileURLTg)
    if err != nil {
        return "", "", err
    }
    defer resp.Body.Close()
    // 3. 保存到本地
    saveName := utils.GenerateUniqueFileName(ext)
    destPath := filepath.Join("public/uploads", saveName)
    if _, err := utils.SaveUploadedFile(resp.Body, destPath); err != nil {
        return "", "", err
    }
    // 4. 记录文件
    repository.InsertFileRecord(sid, destPath, resp.ContentLength)
    // 5. 构建返回URL
    fileURL = "/uploads/" + saveName
    if config.AppConfig.Domain != "" {
        fileURL = "https://" + config.AppConfig.Domain + fileURL
    }
    // 从原始文件名中提取
    // 可通过 Content-Disposition 获取，但这里简单使用
    fileName = filepath.Base(filePath)
    return fileURL, fileName, nil
}
