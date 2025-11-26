package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *EventHandler) TagSuggest(c *gin.Context) {
    var req TagSuggestRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    tags := h.Service.SuggestTags(req.Title, req.Summary, req.Body)

    c.JSON(http.StatusOK, TagSuggestResponse{
        Tags: tags,
    })
}
