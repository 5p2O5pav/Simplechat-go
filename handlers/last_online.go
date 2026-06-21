package handlers

import (
	"encoding/json"
	"net/http"
	"chat-system-go/internal/repository"
)

func LastOnlineHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	if sid == "" {
		http.Error(w, "Missing sid", http.StatusBadRequest)
		return
	}
	sess, err := repository.GetSession(sid)
	if err != nil || sess == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"sid": sid, "lastOnline": nil})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"sid": sid, "lastOnline": sess.LastActive})
}
