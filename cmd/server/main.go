package main

import (
	"events-service/internal/config"
	"events-service/internal/db"
	"events-service/internal/events/handlers"
	"events-service/internal/events/workers"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()
    database := db.InitDB(cfg)

    r := gin.Default()

    r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // React dev servers
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length", "ETag"},
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
}))

    h := handlers.NewEventHandler(database)

    // start broadcast worker
    bw := workers.NewBroadcastWorker(h.Service)
    bw.Start()
    // optionally: store bw to gracefully stop on shutdown


    api := r.Group("/api/v1")
    {
        api.GET("/events", h.ListEvents)
        api.GET("/events/:id", h.GetEvent)
        api.POST("/events", h.CreateEvent)
        api.PATCH("/events/:id", h.UpdateEvent)
        api.POST("/events/:id/moderate", h.ModerateEvent)
        api.POST("/events/:id/broadcast", h.ManualBroadcast)
        api.POST("/events/tag-suggest", h.TagSuggest)
        api.GET("/tags", h.ListTags)
    }

    r.GET("/healthz", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":" + cfg.Port)
}
