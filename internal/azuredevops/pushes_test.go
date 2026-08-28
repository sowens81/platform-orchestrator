package azuredevops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

func TestClient_PushFiles(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			wantPath := "/PlatformEngineering/_apis/git/repositories/repo-123/pushes"

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

			var body pushRequest

			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf(
					"decode push request: %v",
					err,
				)
			}

			if len(body.RefUpdates) != 1 {
				t.Fatalf(
					"refUpdates = %d, want 1",
					len(body.RefUpdates),
				)
			}

			if body.RefUpdates[0].Name != "refs/heads/main" {
				t.Errorf(
					"ref name = %q",
					body.RefUpdates[0].Name,
				)
			}

			if body.RefUpdates[0].OldObjectID != zeroObjectID {
				t.Errorf(
					"oldObjectId = %q",
					body.RefUpdates[0].OldObjectID,
				)
			}

			if len(body.Commits) != 1 {
				t.Fatalf(
					"commits = %d, want 1",
					len(body.Commits),
				)
			}

			if len(body.Commits[0].Changes) != 2 {
				t.Fatalf(
					"changes = %d, want 2",
					len(body.Commits[0].Changes),
				)
			}

			if body.Commits[0].Changes[0].Item.Path != "/README.md" {
				t.Errorf(
					"path = %q",
					body.Commits[0].Changes[0].Item.Path,
				)
			}

			w.WriteHeader(http.StatusCreated)
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

	files := []templatepkg.TemplateFile{
		{
			Path:    "/README.md",
			Content: "# payments-api",
		},
		{
			Path:    "/.azuredevops/azure-pipelines.yml",
			Content: "trigger: main",
		},
	}

	err := client.PushFiles(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"main",
		files,
	)
	if err != nil {
		t.Fatalf(
			"PushFiles() returned unexpected error: %v",
			err,
		)
	}
}

func TestClient_PushFiles_EncodesProjectName(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			want := "/Platform%20Engineering/_apis/git/repositories/repo-123/pushes"

			if r.URL.EscapedPath() != want {
				t.Errorf(
					"path = %q, want %q",
					r.URL.EscapedPath(),
					want,
				)
			}

			w.WriteHeader(http.StatusCreated)
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

	err := client.PushFiles(
		context.Background(),
		"Platform Engineering",
		"repo-123",
		"main",
		[]templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "test",
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"PushFiles() returned unexpected error: %v",
			err,
		)
	}
}

func TestClient_PushFiles_EncodesRepositoryID(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			want := "/PlatformEngineering/_apis/git/repositories/repo%2F123/pushes"

			if r.URL.EscapedPath() != want {
				t.Errorf(
					"path = %q, want %q",
					r.URL.EscapedPath(),
					want,
				)
			}

			w.WriteHeader(http.StatusCreated)
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

	err := client.PushFiles(
		context.Background(),
		"PlatformEngineering",
		"repo/123",
		"main",
		[]templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "test",
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"PushFiles() returned unexpected error: %v",
			err,
		)
	}
}

func TestClient_PushFiles_ReturnsErrorWhenRepositoryIDIsEmpty(
	t *testing.T,
) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	err := client.PushFiles(
		context.Background(),
		"PlatformEngineering",
		"",
		"main",
		[]templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "test",
			},
		},
	)

	if err == nil {
		t.Fatal("PushFiles() expected error")
	}

	if err.Error() != "repository id is required" {
		t.Errorf(
			"error = %q",
			err.Error(),
		)
	}
}

func TestClient_PushFiles_ReturnsErrorWhenBranchIsEmpty(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	err := client.PushFiles(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"",
		[]templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "test",
			},
		},
	)

	if err == nil {
		t.Fatal("PushFiles() expected error")
	}

	if err.Error() != "branch is required" {
		t.Errorf(
			"error = %q",
			err.Error(),
		)
	}
}

func TestClient_PushFiles_ReturnsErrorWhenFilesAreEmpty(t *testing.T) {
	client := NewClient(
		"https://dev.azure.com/example",
		http.DefaultClient,
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	err := client.PushFiles(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"main",
		nil,
	)

	if err == nil {
		t.Fatal("PushFiles() expected error")
	}

	if err.Error() != "at least one file is required" {
		t.Errorf(
			"error = %q",
			err.Error(),
		)
	}
}

func TestClient_PushFiles_ReturnsAzureDevOpsErrorBody(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusBadRequest)

			_, _ = w.Write(
				[]byte(`{"message":"push rejected"}`),
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

	err := client.PushFiles(
		context.Background(),
		"PlatformEngineering",
		"repo-123",
		"main",
		[]templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "test",
			},
		},
	)

	if err == nil {
		t.Fatal("PushFiles() expected error")
	}

	want := `azure devops returned status 400: {"message":"push rejected"}`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}
