package handlers

import (
	"encoding/json"
	"net/http"
	"chat-system-go/internal/config"
	"chat-system-go/internal/services"
	"chat-system-go/internal/repository"
)

func InitHandler(w http.ResponseWriter, r *http.Request) {
	var sid, token string
	// 读取 cookie
	cookie, err := r.Cookie("cw_sid")
	if err == nil && cookie.Value != "" {
		sid = cookie.Value
		// 检查会话是否有效
		valid, err := services.ValidateSession(sid, "")
		if err == nil && valid {
			// 获取token
			sess, _ := repository.GetSession(sid)
			if sess != nil {
				token = sess.Token
			}
		} else {
			sid = ""
		}
	}
	if sid == "" {
		sid, token, err = services.NewSession()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	} else if token == "" {
		// 无效会话，重新生成
		sid, token, err = services.NewSession()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	// 设置 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "cw_sid",
		Value:    sid,
		MaxAge:   180 * 24 * 3600,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		HttpOnly: false,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"sid": sid, "token": token})
}
