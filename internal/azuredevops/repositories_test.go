package azuredevops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sowens81/platform-orchestrator/internal/repository"
)

func TestClient_CreateRepository(t *testing.T) {
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

			wantPath := "/PlatformEngineering/_apis/git/repositories"

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

			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf(
					"Authorization = %q",
					r.Header.Get("Authorization"),
				)
			}

			var body createRepositoryRequest

			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf(
					"decode request: %v",
					err,
				)
			}

			if body.Name != "payments-api" {
				t.Errorf(
					"name = %q, want payments-api",
					body.Name,
				)
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusCreated)

			_, _ = w.Write(
				[]byte(`{
					"id":"repo-123",
					"name":"payments-api",
					"webUrl":"https://dev.azure.com/example/_git/payments-api"
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

	result, err := client.CreateRepository(
		context.Background(),
		"PlatformEngineering",
		"payments-api",
	)
	if err != nil {
		t.Fatalf(
			"CreateRepository() returned unexpected error: %v",
			err,
		)
	}

	if result.ID != "repo-123" {
		t.Errorf(
			"id = %q, want repo-123",
			result.ID,
		)
	}

	if result.Name != "payments-api" {
		t.Errorf(
			"name = %q, want payments-api",
			result.Name,
		)
	}
}

func TestClient_CreateRepository_EncodesProjectName(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			want := "/Platform%20Engineering/_apis/git/repositories"

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

			w.WriteHeader(http.StatusCreated)

			_, _ = w.Write(
				[]byte(`{
					"id":"repo-123",
					"name":"payments-api",
					"webUrl":"https://example"
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

	_, err := client.CreateRepository(
		context.Background(),
		"Platform Engineering",
		"payments-api",
	)
	if err != nil {
		t.Fatalf(
			"CreateRepository() returned unexpected error: %v",
			err,
		)
	}
}

func TestClient_CreateRepository_ReturnsAzureDevOpsErrorBody(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusConflict)

			_, _ = w.Write(
				[]byte(`{"message":"repository already exists"}`),
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

	_, err := client.CreateRepository(
		context.Background(),
		"PlatformEngineering",
		"payments-api",
	)

	if err == nil {
		t.Fatal("CreateRepository() expected error")
	}

	want := `azure devops returned status 409: {"message":"repository already exists"}`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}

func TestClient_CreateRepository_ReturnsErrorWhenResponseHasNoID(
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

			w.WriteHeader(http.StatusCreated)

			_, _ = w.Write(
				[]byte(`{
					"name":"payments-api",
					"webUrl":"https://example"
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

	_, err := client.CreateRepository(
		context.Background(),
		"PlatformEngineering",
		"payments-api",
	)

	if err == nil {
		t.Fatal("CreateRepository() expected error")
	}

	if err.Error() != "create repository response missing id" {
		t.Errorf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestClient_CreateRepository_ReturnsErrorWhenResponseHasNoName(
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

			w.WriteHeader(http.StatusCreated)

			_, _ = w.Write(
				[]byte(`{
					"id":"repo-123",
					"webUrl":"https://example"
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

	_, err := client.CreateRepository(
		context.Background(),
		"PlatformEngineering",
		"payments-api",
	)

	if err == nil {
		t.Fatal("CreateRepository() expected error")
	}

	if err.Error() != "create repository response missing name" {
		t.Errorf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestClient_GetRepository_ReturnsExistingRepository(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.Method != http.MethodGet {
				t.Fatalf(
					"method = %q, want %q",
					r.Method,
					http.MethodGet,
				)
			}

			wantPath := "/platform/_apis/git/repositories/payments-api"

			if r.URL.EscapedPath() != wantPath {
				t.Fatalf(
					"path = %q, want %q",
					r.URL.EscapedPath(),
					wantPath,
				)
			}

			if got := r.URL.Query().Get("api-version"); got != "7.1" {
				t.Errorf(
					"api-version = %q, want %q",
					got,
					"7.1",
				)
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte(`{
					"id": "repo-123",
					"name": "payments-api",
					"webUrl": "https://dev.azure.com/example/platform/_git/payments-api"
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

	got, exists, err := client.GetRepository(
		context.Background(),
		"platform",
		"payments-api",
	)
	if err != nil {
		t.Fatalf(
			"GetRepository() returned unexpected error: %v",
			err,
		)
	}

	if !exists {
		t.Fatal(
			"GetRepository() exists = false, want true",
		)
	}

	if got.ID != "repo-123" {
		t.Errorf(
			"repository ID = %q, want %q",
			got.ID,
			"repo-123",
		)
	}

	if got.Name != "payments-api" {
		t.Errorf(
			"repository name = %q, want %q",
			got.Name,
			"payments-api",
		)
	}

	wantURL := "https://dev.azure.com/example/platform/_git/payments-api"

	if got.WebURL != wantURL {
		t.Errorf(
			"repository URL = %q, want %q",
			got.WebURL,
			wantURL,
		)
	}
}

func TestClient_GetRepository_ReturnsNotFound(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusNotFound)
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

	got, exists, err := client.GetRepository(
		context.Background(),
		"platform",
		"missing-repository",
	)
	if err != nil {
		t.Fatalf(
			"GetRepository() returned unexpected error: %v",
			err,
		)
	}

	if exists {
		t.Fatal(
			"GetRepository() exists = true, want false",
		)
	}

	if got != (repository.Repository{}) {
		t.Errorf(
			"repository = %+v, want zero value",
			got,
		)
	}
}

func TestClient_GetRepository_ReturnsErrorForUnexpectedStatus(
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

			w.WriteHeader(
				http.StatusInternalServerError,
			)

			_, _ = w.Write(
				[]byte(`{
					"message": "internal server error"
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

	_, exists, err := client.GetRepository(
		context.Background(),
		"platform",
		"payments-api",
	)

	if err == nil {
		t.Fatal(
			"GetRepository() error = nil, want error",
		)
	}

	if exists {
		t.Fatal(
			"GetRepository() exists = true, want false",
		)
	}

	expected := `get repository failed: status 500`

	if !strings.Contains(
		err.Error(),
		expected,
	) {
		t.Errorf(
			"error = %q, want to contain %q",
			err,
			expected,
		)
	}
}

func TestClient_BranchExists_ReturnsTrueWhenBranchExists(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf(
						"method = %q, want %q",
						r.Method,
						http.MethodGet,
					)
				}

				expectedPath := "/platform/_apis/git/repositories/repository-id/refs"

				if r.URL.Path != expectedPath {
					t.Errorf(
						"path = %q, want %q",
						r.URL.Path,
						expectedPath,
					)
				}

				if got := r.URL.Query().Get("filter"); got != "heads/main" {
					t.Errorf(
						"filter = %q, want %q",
						got,
						"heads/main",
					)
				}

				if got := r.URL.Query().Get("api-version"); got != "7.1" {
					t.Errorf(
						"api-version = %q, want %q",
						got,
						"7.1",
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{
						"count": 1,
						"value": [
							{
								"name": "refs/heads/main",
								"objectId": "1234567890abcdef"
							}
						]
					}`),
				)
			},
		),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	exists, err := client.BranchExists(
		context.Background(),
		"platform",
		"repository-id",
		"main",
	)
	if err != nil {
		t.Fatalf(
			"BranchExists() returned unexpected error: %v",
			err,
		)
	}

	if !exists {
		t.Error(
			"BranchExists() = false, want true",
		)
	}
}

func TestClient_BranchExists_ReturnsFalseWhenBranchDoesNotExist(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{
						"count": 0,
						"value": []
					}`),
				)
			},
		),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	exists, err := client.BranchExists(
		context.Background(),
		"platform",
		"repository-id",
		"main",
	)
	if err != nil {
		t.Fatalf(
			"BranchExists() returned unexpected error: %v",
			err,
		)
	}

	if exists {
		t.Error(
			"BranchExists() = true, want false",
		)
	}
}

func TestClient_BranchExists_ReturnsErrorForUnexpectedStatus(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusInternalServerError,
				)

				_, _ = w.Write(
					[]byte(`{
						"message": "azure devops unavailable"
					}`),
				)
			},
		),
	)
	defer server.Close()

	client := NewClient(
		server.URL,
		server.Client(),
		&fakeTokenProvider{
			token: "test-token",
		},
	)

	exists, err := client.BranchExists(
		context.Background(),
		"platform",
		"repository-id",
		"main",
	)

	if err == nil {
		t.Fatal(
			"BranchExists() error = nil, want error",
		)
	}

	if exists {
		t.Error(
			"BranchExists() = true, want false",
		)
	}

	expected := "branch lookup failed: status 500"

	if !strings.Contains(
		err.Error(),
		expected,
	) {
		t.Errorf(
			"error = %q, want to contain %q",
			err,
			expected,
		)
	}
}
