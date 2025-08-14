package handlers

import (
	"net/http"
	"provolo-api/internal/types"

	"github.com/gin-gonic/gin"
)

// GetHealthCheck handles the health check request
// @Summary Get health check
// @Description Returns a simple message to indicate the service is up
// @Tags health
// @Produce json
// @Success 200 {object} types.APIResponse
// @Router /api/v1/health [get]
func GetHealthCheck(config types.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.NewSuccessResponse(
			"Health Check",
			"API is running successfully",
			gin.H{
				"uptime":  "running",
				"version": "1.0.0",
				"env":     config.Environment,
				"port":    config.Port,
			},
		)
		c.JSON(http.StatusOK, response)
	}
}
