package repository

import (
	"context"
	"fmt"
	"testing"

	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

func TestService_Create_ReusesExistingRepository(
	t *testing.T,
) {
	templateService := &idempotencyTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path: "/.azuredevops/azure-pipelines.yml",
				Content: `trigger:
- main
`,
			},
		},
	}

	repositoryProvider := &idempotencyRepositoryProvider{
		existingRepository: Repository{
			ID:     "existing-repository-id",
			Name:   "payments-api",
			WebURL: "https://dev.azure.com/example/project/_git/payments-api",
		},
	}

	pipelineProvider := &idempotencyPipelineProvider{
		pipeline: Pipeline{
			ID:   123,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	result, err := service.Create(
		context.Background(),
		CreateRequest{
			Project:        "platform",
			RepositoryName: "payments-api",
			Template:       "dotnet-api",
			Values: map[string]string{
				"SERVICE_NAME": "payments-api",
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"Create() returned unexpected error: %v",
			err,
		)
	}

	if repositoryProvider.createCalls != 0 {
		t.Errorf(
			"CreateRepository() calls = %d, want 0",
			repositoryProvider.createCalls,
		)
	}

	if repositoryProvider.getCalls != 1 {
		t.Errorf(
			"GetRepository() calls = %d, want 1",
			repositoryProvider.getCalls,
		)
	}

	if repositoryProvider.pushCalls != 1 {
		t.Errorf(
			"PushFiles() calls = %d, want 1",
			repositoryProvider.pushCalls,
		)
	}

	if result.Repository.ID != "existing-repository-id" {
		t.Errorf(
			"repository ID = %q, want %q",
			result.Repository.ID,
			"existing-repository-id",
		)
	}
}

type idempotencyTemplateService struct {
	files []templatepkg.TemplateFile
}

func (s *idempotencyTemplateService) Load(
	_ string,
	_ map[string]string,
) ([]templatepkg.TemplateFile, error) {
	return s.files, nil
}

type idempotencyRepositoryProvider struct {
	existingRepository Repository
	getErr             error

	getCalls    int
	createCalls int
	pushCalls   int
}

func (p *idempotencyRepositoryProvider) GetRepository(
	_ context.Context,
	_ string,
	_ string,
) (Repository, bool, error) {
	p.getCalls++

	if p.getErr != nil {
		return Repository{}, false, p.getErr
	}

	return p.existingRepository, true, nil
}

func (p *idempotencyRepositoryProvider) CreateRepository(
	_ context.Context,
	_ string,
	_ string,
) (Repository, error) {
	p.createCalls++

	return Repository{
		ID:     "new-repository-id",
		Name:   "payments-api",
		WebURL: "https://dev.azure.com/example/project/_git/payments-api",
	}, nil
}

func (p *idempotencyRepositoryProvider) PushFiles(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ []templatepkg.TemplateFile,
) error {
	p.pushCalls++
	return nil
}

type idempotencyPipelineProvider struct {
	pipeline    Pipeline
	createCalls int
}

func (p *idempotencyPipelineProvider) CreatePipeline(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
) (Pipeline, error) {
	p.createCalls++
	return p.pipeline, nil
}

func TestService_Create_StopsWhenRepositoryLookupFails(
	t *testing.T,
) {
	templateService := &idempotencyTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path: "/.azuredevops/azure-pipelines.yml",
				Content: `trigger:
- main
`,
			},
		},
	}

	repositoryProvider := &idempotencyRepositoryProvider{
		getErr: fmt.Errorf("azure devops unavailable"),
	}

	pipelineProvider := &idempotencyPipelineProvider{
		pipeline: Pipeline{
			ID:   123,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	_, err := service.Create(
		context.Background(),
		CreateRequest{
			Project:        "platform",
			RepositoryName: "payments-api",
			Template:       "dotnet-api",
			Values: map[string]string{
				"SERVICE_NAME": "payments-api",
			},
		},
	)

	if err == nil {
		t.Fatal(
			"Create() error = nil, want error",
		)
	}

	expected := `get repository "payments-api": azure devops unavailable`

	if err.Error() != expected {
		t.Errorf(
			"error = %q, want %q",
			err,
			expected,
		)
	}

	if repositoryProvider.createCalls != 0 {
		t.Errorf(
			"CreateRepository() calls = %d, want 0",
			repositoryProvider.createCalls,
		)
	}

	if repositoryProvider.pushCalls != 0 {
		t.Errorf(
			"PushFiles() calls = %d, want 0",
			repositoryProvider.pushCalls,
		)
	}

	if pipelineProvider.createCalls != 0 {
		t.Errorf(
			"CreatePipeline() calls = %d, want 0",
			pipelineProvider.createCalls,
		)
	}
}
