package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

const zeroObjectID = "0000000000000000000000000000000000000000"

type pushRequest struct {
	RefUpdates []refUpdate `json:"refUpdates"`
	Commits    []commit    `json:"commits"`
}

type refUpdate struct {
	Name        string `json:"name"`
	OldObjectID string `json:"oldObjectId"`
}

type commit struct {
	Comment string   `json:"comment"`
	Changes []change `json:"changes"`
}

type change struct {
	ChangeType string     `json:"changeType"`
	Item       changeItem `json:"item"`
	NewContent newContent `json:"newContent"`
}

type changeItem struct {
	Path string `json:"path"`
}

type newContent struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

func (c *Client) PushFiles(
	ctx context.Context,
	project string,
	repositoryID string,
	branch string,
	files []templatepkg.TemplateFile,
) error {
	if strings.TrimSpace(repositoryID) == "" {
		return fmt.Errorf(
			"repository id is required",
		)
	}

	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf(
			"branch is required",
		)
	}

	if len(files) == 0 {
		return fmt.Errorf(
			"at least one file is required",
		)
	}

	changes := make(
		[]change,
		0,
		len(files),
	)

	for _, file := range files {
		changes = append(
			changes,
			change{
				ChangeType: "add",
				Item: changeItem{
					Path: file.Path,
				},
				NewContent: newContent{
					Content:     file.Content,
					ContentType: "rawtext",
				},
			},
		)
	}

	payload := pushRequest{
		RefUpdates: []refUpdate{
			{
				Name:        "refs/heads/" + branch,
				OldObjectID: zeroObjectID,
			},
		},
		Commits: []commit{
			{
				Comment: "Initial commit",
				Changes: changes,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf(
			"marshal push request: %w",
			err,
		)
	}

	endpoint := fmt.Sprintf(
		"%s/%s/_apis/git/repositories/%s/pushes?api-version=7.1",
		c.baseURL,
		url.PathEscape(project),
		url.PathEscape(repositoryID),
	)

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"push files request: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if err := validateResponse(
		resp,
		http.StatusCreated,
	); err != nil {
		return err
	}

	return nil
}
