package main

import (
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"chat-system-go/internal/config"
	"chat-system-go/internal/database"
	"chat-system-go/internal/handlers"
	"chat-system-go/internal/services"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	// 启动清理协程
	go func() {
		for {
			time.Sleep(12 * time.Hour)
			if err := services.CleanupExpiredSessions(); err != nil {
				log.Printf("Cleanup error: %v", err)
			}
		}
	}()
	// 启动时执行一次
	if err := services.CleanupExpiredSessions(); err != nil {
		log.Printf("Initial cleanup error: %v", err)
	}
	
	if err := utils.InitGeoIP("GeoLite2-City.mmdb"); err != nil {
 	   log.Println("GeoIP init failed:", err)
	}
	defer utils.CloseGeoIP()

	r := mux.NewRouter()
	// 静态文件
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("public/uploads"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("public")))

	r.HandleFunc("/init", handlers.InitHandler).Methods("GET")
	r.HandleFunc("/send", handlers.SendHandler).Methods("POST")
	r.HandleFunc("/telegram-webhook", handlers.WebhookHandler).Methods("POST")
	r.HandleFunc("/history", handlers.HistoryHandler).Methods("GET")
	r.HandleFunc("/last-online", handlers.LastOnlineHandler).Methods("GET")

	log.Printf("Server starting on port %s", config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, r))
}
