package carstats

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCarEventStatsHandler returns a car's overheat and check-engine episodes
// over the reliability window, bucketed by day, for the dashboard charts.
func GetCarEventStatsHandler(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		carID := c.Param("carID")
		if carID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "car_id is required"})
			return
		}

		stats, err := service.EventStats(carID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
