package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"chat-system-go/internal/repository"
	"chat-system-go/internal/services"
)

func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	token := r.URL.Query().Get("token")
	if sid == "" || token == "" {
		http.Error(w, "Missing sid or token", http.StatusBadRequest)
		return
	}
	valid, err := services.ValidateSession(sid, token)
	if err != nil || !valid {
		http.Error(w, "Invalid session", http.StatusForbidden)
		return
	}
	after := r.URL.Query().Get("after")
	before := r.URL.Query().Get("before")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if i, err := strconv.Atoi(l); err == nil && i > 0 && i <= 100 {
			limit = i
		}
	}
	var msgs []models.Message
	if before != "" {
		b, _ := strconv.ParseInt(before, 10, 64)
		msgs, err = repository.GetMessagesBySidBefore(sid, b, limit+1)
	} else {
		a := int64(0)
		if after != "" {
			a, _ = strconv.ParseInt(after, 10, 64)
		}
		msgs, err = repository.GetMessagesBySidAfter(sid, a, limit+1)
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	hasMore := false
	if len(msgs) > limit {
		hasMore = true
		msgs = msgs[:limit]
	}
	// 确保响应格式与原来一致
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    msgs,
		"hasMore": hasMore,
	})
}
