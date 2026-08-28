package repository

import (
	"context"
	"errors"
	"testing"

	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

type fakeTemplateService struct {
	templateName string
	values       map[string]string
	files        []templatepkg.TemplateFile
	err          error
}

func (f *fakeTemplateService) Load(
	name string,
	values map[string]string,
) ([]templatepkg.TemplateFile, error) {
	f.templateName = name
	f.values = values

	return f.files, f.err
}

type fakeRepositoryProvider struct {
	project        string
	repositoryName string
	repository     Repository
	files          []templatepkg.TemplateFile
	pushedRepoID   string
	pushedBranch   string
	createErr      error
	pushErr        error
}

func (p *fakeRepositoryProvider) GetRepository(
	_ context.Context,
	_ string,
	_ string,
) (Repository, bool, error) {
	return Repository{}, false, nil
}

func (f *fakeRepositoryProvider) CreateRepository(
	ctx context.Context,
	project string,
	name string,
) (Repository, error) {
	f.project = project
	f.repositoryName = name

	return f.repository, f.createErr
}

func (f *fakeRepositoryProvider) PushFiles(
	ctx context.Context,
	project string,
	repositoryID string,
	branch string,
	files []templatepkg.TemplateFile,
) error {
	f.project = project
	f.pushedRepoID = repositoryID
	f.pushedBranch = branch
	f.files = files

	return f.pushErr
}

type fakePipelineProvider struct {
	project      string
	repositoryID string
	pipelineName string
	yamlPath     string
	pipeline     Pipeline
	err          error
}

func (f *fakePipelineProvider) CreatePipeline(
	ctx context.Context,
	project string,
	repositoryID string,
	name string,
	yamlPath string,
) (Pipeline, error) {
	f.project = project
	f.repositoryID = repositoryID
	f.pipelineName = name
	f.yamlPath = yamlPath

	return f.pipeline, f.err
}

func validCreateRequest() CreateRequest {
	return CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}
}

func TestService_Create_LoadsRequestedTemplate(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{
		pipeline: Pipeline{
			ID:   42,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	if templateService.templateName != "dotnet-api" {
		t.Errorf(
			"template name = %q, want %q",
			templateService.templateName,
			"dotnet-api",
		)
	}

	if templateService.values["SERVICE_NAME"] != "payments-api" {
		t.Errorf(
			"SERVICE_NAME = %q, want %q",
			templateService.values["SERVICE_NAME"],
			"payments-api",
		)
	}
}

func TestService_Create_CreatesRepository(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{
		pipeline: Pipeline{
			ID:   42,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	if repositoryProvider.project != "PlatformEngineering" {
		t.Errorf(
			"project = %q, want %q",
			repositoryProvider.project,
			"PlatformEngineering",
		)
	}

	if repositoryProvider.repositoryName != "payments-api" {
		t.Errorf(
			"repository name = %q, want %q",
			repositoryProvider.repositoryName,
			"payments-api",
		)
	}
}

func TestService_Create_PushesRenderedTemplateFiles(t *testing.T) {
	templateFiles := []templatepkg.TemplateFile{
		{
			Path:    "/README.md",
			Content: "# payments-api",
		},
		{
			Path:    "/.azuredevops/azure-pipelines.yml",
			Content: "trigger: main",
		},
	}

	templateService := &fakeTemplateService{
		files: templateFiles,
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{
		pipeline: Pipeline{
			ID:   42,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	if repositoryProvider.pushedRepoID != "repo-123" {
		t.Errorf(
			"repository ID = %q, want %q",
			repositoryProvider.pushedRepoID,
			"repo-123",
		)
	}

	if repositoryProvider.pushedBranch != "main" {
		t.Errorf(
			"branch = %q, want %q",
			repositoryProvider.pushedBranch,
			"main",
		)
	}

	if len(repositoryProvider.files) != 2 {
		t.Fatalf(
			"pushed %d files, want 2",
			len(repositoryProvider.files),
		)
	}

	if repositoryProvider.files[0].Path != "/README.md" {
		t.Errorf(
			"first file path = %q, want %q",
			repositoryProvider.files[0].Path,
			"/README.md",
		)
	}
}

func TestService_Create_CreatesPipeline(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{
		pipeline: Pipeline{
			ID:   42,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	if pipelineProvider.repositoryID != "repo-123" {
		t.Errorf(
			"pipeline repository ID = %q, want %q",
			pipelineProvider.repositoryID,
			"repo-123",
		)
	}

	if pipelineProvider.pipelineName != "payments-api-ci" {
		t.Errorf(
			"pipeline name = %q, want %q",
			pipelineProvider.pipelineName,
			"payments-api-ci",
		)
	}

	if pipelineProvider.yamlPath != "/.azuredevops/azure-pipelines.yml" {
		t.Errorf(
			"pipeline YAML path = %q, want %q",
			pipelineProvider.yamlPath,
			"/.azuredevops/azure-pipelines.yml",
		)
	}
}

func TestService_Create_WrapsTemplateLoadError(t *testing.T) {
	templateService := &fakeTemplateService{
		err: errors.New("unresolved template token: __TEAM_NAME__"),
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `load repository template "dotnet-api": unresolved template token: __TEAM_NAME__`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created after template loading failed")
	}
}

func TestService_Create_WrapsRepositoryCreationError(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		createErr: errors.New("permission denied"),
	}

	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `create repository "payments-api": permission denied`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.pushedRepoID != "" {
		t.Fatal("files were pushed after repository creation failed")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created after repository creation failed")
	}
}

func TestService_Create_WrapsPushFilesError(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
		pushErr: errors.New("push rejected"),
	}

	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `push files to repository "payments-api": push rejected`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created after file push failed")
	}
}

func TestService_Create_WrapsPipelineCreationError(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{
		err: errors.New("pipeline permission denied"),
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `create pipeline "payments-api-ci": pipeline permission denied`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}
}

func TestService_Create_ReturnsRepositoryAndPipelineDetails(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:     "repo-123",
			Name:   "payments-api",
			WebURL: "https://dev.azure.com/example/project/_git/payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{
		pipeline: Pipeline{
			ID:   42,
			Name: "payments-api-ci",
		},
	}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	result, err := service.Create(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	if result.Repository.ID != "repo-123" {
		t.Errorf(
			"repository ID = %q, want %q",
			result.Repository.ID,
			"repo-123",
		)
	}

	if result.Repository.Name != "payments-api" {
		t.Errorf(
			"repository name = %q, want %q",
			result.Repository.Name,
			"payments-api",
		)
	}

	if result.Repository.URL != "https://dev.azure.com/example/project/_git/payments-api" {
		t.Errorf(
			"repository URL = %q, want %q",
			result.Repository.URL,
			"https://dev.azure.com/example/project/_git/payments-api",
		)
	}

	if result.Pipeline.ID != 42 {
		t.Errorf(
			"pipeline ID = %d, want %d",
			result.Pipeline.ID,
			42,
		)
	}

	if result.Pipeline.Name != "payments-api-ci" {
		t.Errorf(
			"pipeline name = %q, want %q",
			result.Pipeline.Name,
			"payments-api-ci",
		)
	}
}

func TestService_Create_ReturnsErrorWhenPipelineYamlIsMissing(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{
		repository: Repository{
			ID:   "repo-123",
			Name: "payments-api",
		},
	}

	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := CreateRequest{
		Project:        "PlatformEngineering",
		RepositoryName: "payments-api",
		Template:       "dotnet-api",
		Values: map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	}

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `pipeline yaml "/.azuredevops/azure-pipelines.yml" not found in template "dotnet-api"`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when pipeline yaml was missing")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when pipeline yaml was missing")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateIsEmpty(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains no files`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created for empty template")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created for empty template")
	}
}

func TestService_Create_ReturnsErrorWhenPipelineYamlIsEmpty(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `pipeline yaml "/.azuredevops/azure-pipelines.yml" is empty`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when pipeline yaml was empty")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when pipeline yaml was empty")
	}
}

func TestService_Create_ReturnsErrorWhenPipelineYamlContainsOnlyWhitespace(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "   \n\t",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `pipeline yaml "/.azuredevops/azure-pipelines.yml" is empty`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when pipeline yaml contained only whitespace")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when pipeline yaml contained only whitespace")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateContainsDuplicatePaths(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/README.md",
				Content: "# duplicate",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains duplicate file path "/README.md"`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when template contained duplicate paths")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when template contained duplicate paths")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateContainsEmptyFilePath(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "",
				Content: "invalid",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains file with empty path`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when template contained an empty file path")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when template contained an empty file path")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateContainsNonAbsoluteFilePath(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains non-absolute file path "README.md"`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when template contained a non-absolute file path")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when template contained a non-absolute file path")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateContainsPathTraversal(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/src/../README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains invalid file path "/src/../README.md"`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when template contained path traversal")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when template contained path traversal")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateContainsNonCanonicalPath(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "//README.md",
				Content: "# payments-api",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains invalid file path "//README.md"`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when template contained a non-canonical path")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when template contained a non-canonical path")
	}
}

func TestService_Create_ReturnsErrorWhenTemplateContainsRootFilePath(t *testing.T) {
	templateService := &fakeTemplateService{
		files: []templatepkg.TemplateFile{
			{
				Path:    "/",
				Content: "invalid",
			},
			{
				Path:    "/.azuredevops/azure-pipelines.yml",
				Content: "trigger: main",
			},
		},
	}

	repositoryProvider := &fakeRepositoryProvider{}
	pipelineProvider := &fakePipelineProvider{}

	service := NewService(
		templateService,
		repositoryProvider,
		pipelineProvider,
	)

	req := validCreateRequest()

	_, err := service.Create(
		context.Background(),
		req,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `template "dotnet-api" contains invalid file path "/"`

	if err.Error() != want {
		t.Errorf(
			"error = %q, want %q",
			err.Error(),
			want,
		)
	}

	if repositoryProvider.repositoryName != "" {
		t.Fatal("repository was created when template contained root file path")
	}

	if pipelineProvider.repositoryID != "" {
		t.Fatal("pipeline was created when template contained root file path")
	}
}
