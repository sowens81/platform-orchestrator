package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sowens81/platform-orchestrator/internal/repository"
)

type RepositoryService interface {
	Create(
		ctx context.Context,
		req repository.CreateRequest,
	) (*repository.CreateResult, error)
}

type RepositoryHandler struct {
	service RepositoryService
}

func NewRepositoryHandler(
	service RepositoryService,
) *RepositoryHandler {
	return &RepositoryHandler{
		service: service,
	}
}

// CreateRepository godoc
//
// @Summary Create a repository
// @Description Creates an Azure DevOps repository from a template, pushes the rendered files, and creates the associated YAML pipeline.
// @Tags repositories
// @Accept json
// @Produce json
// @Param request body repository.CreateRequest true "Repository creation request"
// @Success 201 {object} repository.CreateResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/repositories [post]
func (h *RepositoryHandler) Create(
	c *gin.Context,
) {
	var req repository.CreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			ErrorResponse{
				Error: "invalid request body",
			},
		)

		return
	}

	if err := validateCreateRepositoryRequest(req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			ErrorResponse{
				Error: err.Error(),
			},
		)

		return
	}

	result, err := h.service.Create(
		c.Request.Context(),
		req,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			ErrorResponse{
				Error: err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusCreated,
		result,
	)
}

func validateCreateRepositoryRequest(
	req repository.CreateRequest,
) error {
	if req.Project == "" {
		return fmt.Errorf(
			"project is required",
		)
	}

	if req.RepositoryName == "" {
		return fmt.Errorf(
			"repositoryName is required",
		)
	}

	if req.Template == "" {
		return fmt.Errorf(
			"template is required",
		)
	}

	if len(req.Values) == 0 {
		return fmt.Errorf(
			"values is required",
		)
	}

	return nil
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}
