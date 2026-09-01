package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sowens81/platform-orchestrator/internal/repository"
)

type fakeRepositoryService struct {
	request repository.CreateRequest
	result  *repository.CreateResult
	err     error

	called bool
}

func (f *fakeRepositoryService) Create(
	ctx context.Context,
	req repository.CreateRequest,
) (*repository.CreateResult, error) {
	f.called = true
	f.request = req

	return f.result, f.err
}

func validRequestBody() []byte {
	return []byte(`{
		"project":"PlatformEngineering",
		"repositoryName":"payments-api",
		"template":"dotnet-api",
		"values":{
			"SERVICE_NAME":"payments-api"
		}
	}`)
}

func validResult() *repository.CreateResult {
	return &repository.CreateResult{
		Repository: repository.RepositoryResult{
			ID:   "repo-123",
			Name: "payments-api",
			URL:  "https://dev.azure.com/example/_git/payments-api",
		},
		Pipeline: repository.PipelineResult{
			ID:   42,
			Name: "payments-api-build",
		},
	}
}

func performRepositoryRequest(
	handler *RepositoryHandler,
	body []byte,
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	// Preserve the strict JSON behaviour from the previous net/http
	// implementation which used json.Decoder.DisallowUnknownFields().
	gin.EnableJsonDecoderDisallowUnknownFields()

	router := gin.New()

	router.POST(
		"/v1/repositories",
		handler.Create,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/repositories",
		bytes.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	return recorder
}

func TestRepositoryHandler_Create_ReturnsCreated(t *testing.T) {
	service := &fakeRepositoryService{
		result: validResult(),
	}

	handler := NewRepositoryHandler(service)

	recorder := performRepositoryRequest(
		handler,
		validRequestBody(),
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}

	contentType := recorder.Header().Get(
		"Content-Type",
	)

	if !strings.HasPrefix(
		contentType,
		"application/json",
	) {
		t.Errorf(
			"Content-Type = %q, want application/json",
			contentType,
		)
	}

	var result repository.CreateResult

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&result); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if result.Repository.ID != "repo-123" {
		t.Errorf(
			"repository id = %q",
			result.Repository.ID,
		)
	}

	if result.Pipeline.ID != 42 {
		t.Errorf(
			"pipeline id = %d",
			result.Pipeline.ID,
		)
	}
}

func TestRepositoryHandler_Create_PassesRequestToService(
	t *testing.T,
) {
	service := &fakeRepositoryService{
		result: validResult(),
	}

	handler := NewRepositoryHandler(service)

	recorder := performRepositoryRequest(
		handler,
		validRequestBody(),
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}

	if service.request.Project != "PlatformEngineering" {
		t.Errorf(
			"project = %q",
			service.request.Project,
		)
	}

	if service.request.RepositoryName != "payments-api" {
		t.Errorf(
			"repositoryName = %q",
			service.request.RepositoryName,
		)
	}

	if service.request.Template != "dotnet-api" {
		t.Errorf(
			"template = %q",
			service.request.Template,
		)
	}

	if service.request.Values["SERVICE_NAME"] != "payments-api" {
		t.Errorf(
			"SERVICE_NAME = %q",
			service.request.Values["SERVICE_NAME"],
		)
	}
}

func TestRepositoryHandler_Create_ReturnsBadRequestForInvalidJSON(
	t *testing.T,
) {
	service := &fakeRepositoryService{}

	handler := NewRepositoryHandler(service)

	recorder := performRepositoryRequest(
		handler,
		[]byte("{"),
	)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}

	if service.called {
		t.Error(
			"service should not have been called",
		)
	}
}

func TestRepositoryHandler_Create_RejectsUnknownFields(
	t *testing.T,
) {
	service := &fakeRepositoryService{}

	handler := NewRepositoryHandler(service)

	body := []byte(`{
		"project":"PlatformEngineering",
		"repositoryName":"payments-api",
		"template":"dotnet-api",
		"unknown":"value",
		"values":{
			"SERVICE_NAME":"payments-api"
		}
	}`)

	recorder := performRepositoryRequest(
		handler,
		body,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}

	if service.called {
		t.Error(
			"service should not have been called",
		)
	}
}

func TestRepositoryHandler_Create_RequiresProject(
	t *testing.T,
) {
	testValidationError(
		t,
		`{
			"repositoryName":"payments-api",
			"template":"dotnet-api",
			"values":{
				"SERVICE_NAME":"payments-api"
			}
		}`,
		"project is required",
	)
}

func TestRepositoryHandler_Create_RequiresRepositoryName(
	t *testing.T,
) {
	testValidationError(
		t,
		`{
			"project":"PlatformEngineering",
			"template":"dotnet-api",
			"values":{
				"SERVICE_NAME":"payments-api"
			}
		}`,
		"repositoryName is required",
	)
}

func TestRepositoryHandler_Create_RequiresTemplate(
	t *testing.T,
) {
	testValidationError(
		t,
		`{
			"project":"PlatformEngineering",
			"repositoryName":"payments-api",
			"values":{
				"SERVICE_NAME":"payments-api"
			}
		}`,
		"template is required",
	)
}

func TestRepositoryHandler_Create_RequiresValues(
	t *testing.T,
) {
	testValidationError(
		t,
		`{
			"project":"PlatformEngineering",
			"repositoryName":"payments-api",
			"template":"dotnet-api",
			"values":{}
		}`,
		"values is required",
	)
}

func TestRepositoryHandler_Create_ReturnsInternalServerErrorWhenServiceFails(
	t *testing.T,
) {
	service := &fakeRepositoryService{
		err: errors.New(
			"repository creation failed",
		),
	}

	handler := NewRepositoryHandler(service)

	recorder := performRepositoryRequest(
		handler,
		validRequestBody(),
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}

	var response ErrorResponse

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode error response: %v",
			err,
		)
	}

	if response.Error != "repository creation failed" {
		t.Errorf(
			"error = %q",
			response.Error,
		)
	}
}

func testValidationError(
	t *testing.T,
	body string,
	expectedError string,
) {
	t.Helper()

	service := &fakeRepositoryService{}

	handler := NewRepositoryHandler(service)

	recorder := performRepositoryRequest(
		handler,
		[]byte(body),
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}

	var response ErrorResponse

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if response.Error != expectedError {
		t.Errorf(
			"error = %q, want %q",
			response.Error,
			expectedError,
		)
	}

	if service.called {
		t.Error(
			"service should not have been called",
		)
	}
}
