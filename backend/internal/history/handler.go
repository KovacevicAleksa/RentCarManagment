package history

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type HistoryQueryParams struct {
	Limit int       `form:"limit"`
	Start time.Time `form:"start" time_format:"2006-01-02T15:04:05Z07:00"`
	End   time.Time `form:"end" time_format:"2006-01-02T15:04:05Z07:00"`
}

func GetCarHistoryHandler(service *HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		carID := c.Param("carID")
		if carID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "car_id is required"})
			return
		}

		var params HistoryQueryParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var records []CarHistory
		var err error

		if !params.Start.IsZero() && !params.End.IsZero() {
			records, err = service.GetHistoryInRange(carID, params.Start, params.End)
		} else {
			limit := params.Limit
			if limit == 0 {
				limit = 20 // Promenjen default sa 100 na 20
			}
			records, err = service.GetRecentHistory(carID, limit)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"car_id":  carID,
			"count":   len(records),
			"records": records,
		})
	}
}

func GetLatestHandler(service *HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		carID := c.Param("carID")
		if carID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "car_id is required"})
			return
		}

		record, err := service.GetLatest(carID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no records found"})
			return
		}

		c.JSON(http.StatusOK, record)
	}
}

func CleanupHandler(service *HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.Query("days")
		if daysStr == "" {
			daysStr = "30"
		}

		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter"})
			return
		}

		if err := service.CleanupOldRecords(days); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "cleanup successful",
			"days":    days,
		})
	}
}