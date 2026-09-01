package azuredevops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreatePipeline(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.Method != http.MethodPost {
				t.Errorf(
					"method = %q, want POST",
					r.Method,
				)
			}

			wantPath := "/PlatformEngineering/_apis/pipelines"

			if r.URL.Path != wantPath {
				t.Errorf(
					"path = %q, want %q",
					r.URL.Path,
					wantPath,
				)
			}

			if r.URL.Query().Get("api-version") != "7.1" {
				t.Errorf(
					"api-version = %q, want 7.1",
					r.URL.Query().Get("api-version"),
				)
			}

			var body createPipelineRequest

			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf(
					"decode request: %v",
					err,
				)
			}

			if body.Name != "payments-api-build" {
				t.Errorf(
					"name = %q",
					body.Name,
				)
			}

			if body.Configuration.Type != "yaml" {
				t.Errorf(
					"type = %q",
					body.Configuration.Type,
				)
			}

			if body.Configuration.Path != "/.azuredevops/azure-pipelines.yml" {
				t.Errorf(
					"path = %q",
					body.Configuration.Path,
				)
			}

			if body.Configuration.Repository.ID != "repo-123" {
				t.Errorf(
					"repository id = %q",
					body.Configuration.Repository.ID,
				)
			}

			if body.Configuration.Repository.Type != "azureReposGit" {
				t.Errorf(
					"repository type = %q",
					body.Configuration.Repository.Type,
				)
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte(`{
					"id":42,
					"name":"payments-api-build"
				}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	result, err := client.CreatePipeline(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"payments-api-build",
		"/.azuredevops/azure-pipelines.yml",
	)
	if err != nil {
		t.Fatalf(
			"CreatePipeline() returned unexpected error: %v",
			err,
		)
	}

	if result.ID != 42 {
		t.Errorf(
			"id = %d, want 42",
			result.ID,
		)
	}

	if result.Name != "payments-api-build" {
		t.Errorf(
			"name = %q",
			result.Name,
		)
	}
}

func TestClient_CreatePipeline_UsesStableAPIVersion(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.URL.Query().Get("api-version") != "7.1" {
				t.Errorf(
					"api-version = %q, want 7.1",
					r.URL.Query().Get("api-version"),
				)
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte(`{
					"id":42,
					"name":"payments-api-build"
				}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	_, err := client.CreatePipeline(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"payments-api-build",
		"/.azuredevops/azure-pipelines.yml",
	)
	if err != nil {
		t.Fatalf(
			"CreatePipeline() returned unexpected error: %v",
			err,
		)
	}
}

func TestClient_CreatePipeline_EncodesProjectName(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			want := "/Platform%20Engineering/_apis/pipelines"

			if r.URL.EscapedPath() != want {
				t.Errorf(
					"path = %q, want %q",
					r.URL.EscapedPath(),
					want,
				)
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte(`{
					"id":42,
					"name":"payments-api-build"
				}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	_, err := client.CreatePipeline(
		context.Background(),
		"Platform Engineering",
		"repo-123",
		"payments-api-build",
		"/.azuredevops/azure-pipelines.yml",
	)
	if err != nil {
		t.Fatalf(
			"CreatePipeline() returned unexpected error: %v",
			err,
		)
	}
}

func TestClient_CreatePipeline_ReturnsAzureDevOpsErrorBody(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusBadRequest)

			_, _ = w.Write(
				[]byte(`{"message":"pipeline rejected"}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	_, err := client.CreatePipeline(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"payments-api-build",
		"/.azuredevops/azure-pipelines.yml",
	)

	if err == nil {
		t.Fatal("CreatePipeline() expected error")
	}

	want := `azure devops returned status 400: {"message":"pipeline rejected"}`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}

func TestClient_CreatePipeline_ReturnsErrorWhenResponseHasNoID(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte(`{
					"name":"payments-api-build"
				}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	_, err := client.CreatePipeline(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"payments-api-build",
		"/.azuredevops/azure-pipelines.yml",
	)

	if err == nil {
		t.Fatal("CreatePipeline() expected error")
	}

	if err.Error() != "create pipeline response missing id" {
		t.Errorf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestClient_CreatePipeline_ReturnsErrorWhenResponseHasNoName(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte(`{
					"id":42
				}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	_, err := client.CreatePipeline(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"payments-api-build",
		"/.azuredevops/azure-pipelines.yml",
	)

	if err == nil {
		t.Fatal("CreatePipeline() expected error")
	}

	if err.Error() != "create pipeline response missing name" {
		t.Errorf(
			"unexpected error: %v",
			err,
		)
	}
}
