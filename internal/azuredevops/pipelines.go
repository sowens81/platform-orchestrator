package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sowens81/platform-orchestrator/internal/repository"
)

type createPipelineRequest struct {
	Name          string                `json:"name"`
	Configuration pipelineConfiguration `json:"configuration"`
}

type pipelineConfiguration struct {
	Type       string             `json:"type"`
	Path       string             `json:"path"`
	Repository pipelineRepository `json:"repository"`
}

type pipelineRepository struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type createPipelineResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c *Client) CreatePipeline(
	ctx context.Context,
	project string,
	repositoryID string,
	name string,
	yamlPath string,
) (repository.Pipeline, error) {
	payload := createPipelineRequest{
		Name: name,
		Configuration: pipelineConfiguration{
			Type: "yaml",
			Path: yamlPath,
			Repository: pipelineRepository{
				ID:   repositoryID,
				Type: "azureReposGit",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return repository.Pipeline{}, fmt.Errorf(
			"marshal create pipeline request: %w",
			err,
		)
	}

	endpoint := fmt.Sprintf(
		"%s/%s/_apis/pipelines?api-version=7.1",
		c.baseURL,
		url.PathEscape(project),
	)

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return repository.Pipeline{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return repository.Pipeline{}, fmt.Errorf(
			"create pipeline request: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if err := validateResponse(
		resp,
		http.StatusOK,
	); err != nil {
		return repository.Pipeline{}, err
	}

	var result createPipelineResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return repository.Pipeline{}, fmt.Errorf(
			"decode create pipeline response: %w",
			err,
		)
	}

	if result.ID == 0 {
		return repository.Pipeline{}, fmt.Errorf(
			"create pipeline response missing id",
		)
	}

	if result.Name == "" {
		return repository.Pipeline{}, fmt.Errorf(
			"create pipeline response missing name",
		)
	}

	return repository.Pipeline{
		ID:   result.ID,
		Name: result.Name,
	}, nil
}

var _ repository.PipelineProvider = (*Client)(nil)
