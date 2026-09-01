package repository

import (
	"context"
	"testing"

	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

func TestServiceRegistration_StoresProvisionedResourceIdentity(
	t *testing.T,
) {
	registration := ServiceRegistration{
		ServiceName: "payments-api",
		Project:     "platform",

		RepositoryID:   "repo-123",
		RepositoryName: "payments-api",

		BuildPipelineID:   100,
		BuildPipelineName: "payments-api-build",

		ReleasePipelineID:   200,
		ReleasePipelineName: "payments-api-release",

		TemplateName:    "dotnet-api",
		TemplateVersion: "1.0.0",
	}

	if registration.ServiceName != "payments-api" {
		t.Errorf(
			"ServiceName = %q, want %q",
			registration.ServiceName,
			"payments-api",
		)
	}

	if registration.RepositoryID != "repo-123" {
		t.Errorf(
			"RepositoryID = %q, want %q",
			registration.RepositoryID,
			"repo-123",
		)
	}

	if registration.BuildPipelineID != 100 {
		t.Errorf(
			"BuildPipelineID = %d, want %d",
			registration.BuildPipelineID,
			100,
		)
	}

	if registration.ReleasePipelineID != 200 {
		t.Errorf(
			"ReleasePipelineID = %d, want %d",
			registration.ReleasePipelineID,
			200,
		)
	}
}

func TestServiceRegistrationStore_Contract(
	t *testing.T,
) {
	var _ ServiceRegistrationStore = (*fakeServiceRegistrationStore)(nil)
}

type fakeServiceRegistrationStore struct {
	registration ServiceRegistration
	exists       bool
}

func (s *fakeServiceRegistrationStore) Get(
	_ context.Context,
	_ string,
	_ string,
) (ServiceRegistration, bool, error) {
	return s.registration, s.exists, nil
}

func (s *fakeServiceRegistrationStore) Save(
	_ context.Context,
	registration ServiceRegistration,
) error {
	s.registration = registration
	s.exists = true

	return nil
}

func TestService_Create_ReusesRepositoryFromServiceRegistration(
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

	registrationStore := &fakeServiceRegistrationStore{
		exists: true,
		registration: ServiceRegistration{
			ServiceName: "payments-api",
			Project:     "platform",

			RepositoryID:   "registered-repository-id",
			RepositoryName: "payments-api",

			BuildPipelineID:   123,
			BuildPipelineName: "payments-api-build",

			ReleasePipelineID:   456,
			ReleasePipelineName: "payments-api-release",

			TemplateName:    "dotnet-api",
			TemplateVersion: "1.0.0",
		},
	}

	repositoryProvider := &idempotencyRepositoryProvider{
		existingRepository: Repository{
			ID:     "discovered-repository-id",
			Name:   "payments-api",
			WebURL: "https://dev.azure.com/example/platform/_git/payments-api",
		},
		branchExists: true,
	}

	pipelineProvider := &idempotencyPipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
		registrationStore,
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

	if repositoryProvider.getCalls != 0 {
		t.Errorf(
			"GetRepository() calls = %d, want 0",
			repositoryProvider.getCalls,
		)
	}

	if result.Repository.ID != "registered-repository-id" {
		t.Errorf(
			"repository ID = %q, want %q",
			result.Repository.ID,
			"registered-repository-id",
		)
	}
}

func TestService_Create_ReusesBuildPipelineFromServiceRegistration(
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

	registrationStore := &fakeServiceRegistrationStore{
		exists: true,
		registration: ServiceRegistration{
			ServiceName: "payments-api",
			Project:     "platform",

			RepositoryID:   "repo-123",
			RepositoryName: "payments-api",
			RepositoryURL:  "https://dev.azure.com/example/platform/_git/payments-api",

			BuildPipelineID:   100,
			BuildPipelineName: "payments-api-build",

			TemplateName:    "dotnet-api",
			TemplateVersion: "1.0.0",
		},
	}

	repositoryProvider := &idempotencyRepositoryProvider{
		branchExists: true,
	}

	pipelineProvider := &idempotencyPipelineProvider{
		pipeline: Pipeline{
			ID:   999,
			Name: "should-not-be-created",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
		registrationStore,
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

	if pipelineProvider.createCalls != 0 {
		t.Errorf(
			"CreatePipeline() calls = %d, want 0",
			pipelineProvider.createCalls,
		)
	}

	if result.Pipeline.ID != 100 {
		t.Errorf(
			"pipeline ID = %d, want %d",
			result.Pipeline.ID,
			100,
		)
	}

	if result.Pipeline.Name != "payments-api-build" {
		t.Errorf(
			"pipeline name = %q, want %q",
			result.Pipeline.Name,
			"payments-api-build",
		)
	}
}
