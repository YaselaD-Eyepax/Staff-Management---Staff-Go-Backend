package main

import (
	"events-service/internal/config"
	"events-service/internal/db"
	"events-service/internal/events/handlers"
	"events-service/internal/events/workers"

	"github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()
    database := db.InitDB(cfg)

    r := gin.Default()

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
        // api.POST("/events/:id/broadcast", h.BroadcastEvent)
        // api.POST("/events/tag-suggest", h.TagSuggest)
        // api.GET("/tags", h.ListTags)
    }

    r.GET("/healthz", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":" + cfg.Port)
}
