package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *EventHandler) ListTags(c *gin.Context) {
    q := c.Query("query")

    tags, err := h.Service.SearchGlobalTags(q)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load tags"})
        return
    }

    // Extract tag strings only
    tagList := make([]string, 0, len(tags))
    for _, t := range tags {
        tagList = append(tagList, t.Tag)
    }

    c.JSON(http.StatusOK, gin.H{"tags": tagList})
}
