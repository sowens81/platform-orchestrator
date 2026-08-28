package azuredevops

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeTokenProvider struct {
	token string
	err   error
}

func (f *fakeTokenProvider) Token(
	ctx context.Context,
) (string, error) {
	return f.token, f.err
}

func TestClient_NewRequest_AddsAuthorizationHeader(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	req, err := client.newRequest(
		context.Background(),
		http.MethodPost,
		"https://dev.azure.com/example/test",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"newRequest() returned unexpected error: %v",
			err,
		)
	}

	want := "Bearer test-token"

	if req.Header.Get("Authorization") != want {
		t.Errorf(
			"Authorization = %q, want %q",
			req.Header.Get("Authorization"),
			want,
		)
	}
}

func TestClient_NewRequest_AddsJSONHeaders(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	req, err := client.newRequest(
		context.Background(),
		http.MethodPost,
		"https://dev.azure.com/example/test",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"newRequest() returned unexpected error: %v",
			err,
		)
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Errorf(
			"Accept = %q",
			req.Header.Get("Accept"),
		)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf(
			"Content-Type = %q",
			req.Header.Get("Content-Type"),
		)
	}
}

func TestClient_NewRequest_WrapsTokenProviderError(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{
			err: errors.New("token acquisition failed"),
		},
	)

	_, err := client.newRequest(
		context.Background(),
		http.MethodGet,
		"https://dev.azure.com/example/test",
		nil,
	)

	if err == nil {
		t.Fatal("newRequest() expected error")
	}

	want := "acquire azure devops token: token acquisition failed"

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}

func TestClient_NewRequest_ReturnsErrorWhenTokenIsEmpty(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{},
	)

	_, err := client.newRequest(
		context.Background(),
		http.MethodGet,
		"https://dev.azure.com/example/test",
		nil,
	)

	if err == nil {
		t.Fatal("newRequest() expected error")
	}

	want := "acquire azure devops token: token is empty"

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}

func TestNewClient_RemovesTrailingSlashFromBaseURL(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example/",
		http.DefaultClient,
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	want := "https://dev.azure.com/example"

	if client.baseURL != want {
		t.Errorf(
			"baseURL = %q, want %q",
			client.baseURL,
			want,
		)
	}
}

func TestValidateResponse_AcceptsAllowedStatus(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := validateResponse(
		resp,
		http.StatusOK,
		http.StatusCreated,
	)
	if err != nil {
		t.Fatalf(
			"validateResponse() returned unexpected error: %v",
			err,
		)
	}
}

func TestValidateResponse_ReturnsResponseBodyOnError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body: io.NopCloser(
			strings.NewReader(`{"message":"invalid request"}`),
		),
	}

	err := validateResponse(
		resp,
		http.StatusCreated,
	)

	if err == nil {
		t.Fatal("validateResponse() expected error")
	}

	want := `azure devops returned status 400: {"message":"invalid request"}`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}

func TestValidateResponse_ReturnsStatusWhenBodyIsEmpty(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := validateResponse(
		resp,
		http.StatusCreated,
	)

	if err == nil {
		t.Fatal("validateResponse() expected error")
	}

	want := "azure devops returned status 500"

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}
