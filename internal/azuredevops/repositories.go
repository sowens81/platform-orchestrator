package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sowens81/platform-orchestrator/internal/repository"
)

type createRepositoryRequest struct {
	Name string `json:"name"`
}

type createRepositoryResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	WebURL string `json:"webUrl"`
}

func (c *Client) GetRepository(
	ctx context.Context,
	project string,
	name string,
) (repository.Repository, bool, error) {
	token, err := c.tokenProvider.Token(ctx)
	if err != nil {
		return repository.Repository{}, false, fmt.Errorf(
			"get access token: %w",
			err,
		)
	}

	requestURL := fmt.Sprintf(
		"%s/%s/_apis/git/repositories/%s?api-version=7.1",
		c.baseURL,
		url.PathEscape(project),
		url.PathEscape(name),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)
	if err != nil {
		return repository.Repository{}, false, fmt.Errorf(
			"create get repository request: %w",
			err,
		)
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)
	req.Header.Set(
		"Accept",
		"application/json",
	)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return repository.Repository{}, false, fmt.Errorf(
			"get repository request: %w",
			err,
		)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return repository.Repository{}, false, nil
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)

		return repository.Repository{}, false, fmt.Errorf(
			"get repository failed: status %d: %s",
			res.StatusCode,
			string(body),
		)
	}

	var response struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		WebURL string `json:"webUrl"`
	}

	if err := json.NewDecoder(res.Body).Decode(
		&response,
	); err != nil {
		return repository.Repository{}, false, fmt.Errorf(
			"decode repository response: %w",
			err,
		)
	}

	return repository.Repository{
		ID:     response.ID,
		Name:   response.Name,
		WebURL: response.WebURL,
	}, true, nil
}

func (c *Client) CreateRepository(
	ctx context.Context,
	project string,
	name string,
) (repository.Repository, error) {
	payload, err := json.Marshal(
		createRepositoryRequest{
			Name: name,
		},
	)
	if err != nil {
		return repository.Repository{}, fmt.Errorf(
			"marshal create repository request: %w",
			err,
		)
	}

	endpoint := fmt.Sprintf(
		"%s/%s/_apis/git/repositories?api-version=7.1",
		c.baseURL,
		url.PathEscape(project),
	)

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return repository.Repository{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return repository.Repository{}, fmt.Errorf(
			"create repository request: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if err := validateResponse(
		resp,
		http.StatusCreated,
	); err != nil {
		return repository.Repository{}, err
	}

	var result createRepositoryResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return repository.Repository{}, fmt.Errorf(
			"decode create repository response: %w",
			err,
		)
	}

	if result.ID == "" {
		return repository.Repository{}, fmt.Errorf(
			"create repository response missing id",
		)
	}

	if result.Name == "" {
		return repository.Repository{}, fmt.Errorf(
			"create repository response missing name",
		)
	}

	return repository.Repository{
		ID:     result.ID,
		Name:   result.Name,
		WebURL: result.WebURL,
	}, nil
}

func (c *Client) BranchExists(
	ctx context.Context,
	project string,
	repositoryID string,
	branch string,
) (bool, error) {
	token, err := c.tokenProvider.Token(ctx)
	if err != nil {
		return false, fmt.Errorf(
			"get access token: %w",
			err,
		)
	}

	requestURL := fmt.Sprintf(
		"%s/%s/_apis/git/repositories/%s/refs",
		c.baseURL,
		url.PathEscape(project),
		url.PathEscape(repositoryID),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf(
			"create branch lookup request: %w",
			err,
		)
	}

	query := req.URL.Query()
	query.Set(
		"filter",
		"heads/"+branch,
	)
	query.Set(
		"api-version",
		"7.1",
	)
	req.URL.RawQuery = query.Encode()

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)
	req.Header.Set(
		"Accept",
		"application/json",
	)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf(
			"branch lookup request: %w",
			err,
		)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)

		return false, fmt.Errorf(
			"branch lookup failed: status %d: %s",
			res.StatusCode,
			string(body),
		)
	}

	var response struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return false, fmt.Errorf(
			"decode branch lookup response: %w",
			err,
		)
	}

	expectedRef := "refs/heads/" + branch

	for _, ref := range response.Value {
		if ref.Name == expectedRef {
			return true, nil
		}
	}

	return false, nil
}

var _ repository.RepositoryProvider = (*Client)(nil)
