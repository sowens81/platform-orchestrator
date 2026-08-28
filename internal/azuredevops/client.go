package azuredevops

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

type Client struct {
	baseURL       string
	httpClient    *http.Client
	tokenProvider TokenProvider
}

func NewClient(
	baseURL string,
	httpClient *http.Client,
	tokenProvider TokenProvider,
) *Client {
	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		httpClient:    httpClient,
		tokenProvider: tokenProvider,
	}
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body io.Reader,
) (*http.Request, error) {
	token, err := c.tokenProvider.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"acquire azure devops token: %w",
			err,
		)
	}

	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf(
			"acquire azure devops token: token is empty",
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		endpoint,
		body,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create azure devops request: %w",
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

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	return req, nil
}

func validateResponse(
	resp *http.Response,
	allowedStatusCodes ...int,
) error {
	for _, allowed := range allowedStatusCodes {
		if resp.StatusCode == allowed {
			return nil
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(
			"azure devops returned status %d",
			resp.StatusCode,
		)
	}

	message := strings.TrimSpace(string(body))

	if message == "" {
		return fmt.Errorf(
			"azure devops returned status %d",
			resp.StatusCode,
		)
	}

	return fmt.Errorf(
		"azure devops returned status %d: %s",
		resp.StatusCode,
		message,
	)
}
