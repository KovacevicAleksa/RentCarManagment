package notification

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ListHandler(s *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := s.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list notifications"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"notifications": items})
	}
}

func UnreadCountHandler(s *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := s.UnreadCount()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count notifications"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"unread": count})
	}
}

func MarkReadHandler(s *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.MarkAllRead(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications read"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "All notifications marked read"})
	}
}

type CreateNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type"`
}

func AdminCreateHandler(s *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateNotificationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ntype := req.Type
		if ntype != TypeAlert {
			ntype = TypeInfo
		}

		n, err := s.Create(strings.TrimSpace(req.Title), strings.TrimSpace(req.Message), ntype, SourceAdmin, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"notification": n})
	}
}
