package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status" example:"healthy"`
}

// Health godoc
//
// @Summary Health check
// @Description Returns the current health of the Platform Orchestrator API.
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func Health(
	c *gin.Context,
) {
	c.JSON(
		http.StatusOK,
		HealthResponse{
			Status: "healthy",
		},
	)
}
