package repository

import (
	"context"
	"fmt"
	"path"
	"strings"

	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

const (
	defaultBranch    = "main"
	pipelineYamlPath = "/.azuredevops/azure-pipelines.yml"
)

type TemplateService interface {
	Load(
		name string,
		values map[string]string,
	) (templatepkg.RenderedTemplate, error)
}

type RepositoryProvider interface {
	GetRepository(
		ctx context.Context,
		project string,
		name string,
	) (Repository, bool, error)

	CreateRepository(
		ctx context.Context,
		project string,
		name string,
	) (Repository, error)

	BranchExists(
		ctx context.Context,
		project string,
		repositoryID string,
		branch string,
	) (bool, error)

	PushFiles(
		ctx context.Context,
		project string,
		repositoryID string,
		branch string,
		files []templatepkg.TemplateFile,
	) error
}

type PipelineProvider interface {
	CreatePipeline(
		ctx context.Context,
		project string,
		repositoryID string,
		name string,
		yamlPath string,
	) (Pipeline, error)
}

type Service struct {
	templates         TemplateService
	repositories      RepositoryProvider
	pipelines         PipelineProvider
	registrationStore ServiceRegistrationStore
}

func NewService(
	templates TemplateService,
	repositories RepositoryProvider,
	pipelines PipelineProvider,
	registrationStore ServiceRegistrationStore,
) *Service {
	return &Service{
		templates:         templates,
		repositories:      repositories,
		pipelines:         pipelines,
		registrationStore: registrationStore,
	}
}

func (s *Service) Create(
	ctx context.Context,
	req CreateRequest,
) (*CreateResult, error) {
	rendered, err := s.templates.Load(
		req.Template,
		req.Values,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load template %q: %w",
			req.Template,
			err,
		)
	}

	files := rendered.Files

	if len(files) == 0 {
		return nil, fmt.Errorf(
			"template %q contains no files",
			req.Template,
		)
	}

	if hasEmptyFilePath(files) {
		return nil, fmt.Errorf(
			"template %q contains file with empty path",
			req.Template,
		)
	}

	if nonAbsolutePath, found := findNonAbsolutePath(files); found {
		return nil, fmt.Errorf(
			"template %q contains non-absolute file path %q",
			req.Template,
			nonAbsolutePath,
		)
	}

	if invalidPath, found := findInvalidPath(files); found {
		return nil, fmt.Errorf(
			"template %q contains invalid file path %q",
			req.Template,
			invalidPath,
		)
	}

	if duplicatePath, found := findDuplicatePath(files); found {
		return nil, fmt.Errorf(
			"template %q contains duplicate file path %q",
			req.Template,
			duplicatePath,
		)
	}

	pipelineFile, found := findFile(
		files,
		pipelineYamlPath,
	)
	if !found {
		return nil, fmt.Errorf(
			"pipeline yaml %q not found in template %q",
			pipelineYamlPath,
			req.Template,
		)
	}

	if strings.TrimSpace(pipelineFile.Content) == "" {
		return nil, fmt.Errorf(
			"pipeline yaml %q is empty",
			pipelineYamlPath,
		)
	}

	registration, registrationExists, err := s.registrationStore.Get(
		ctx,
		req.Project,
		req.RepositoryName,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get service registration %q: %w",
			req.RepositoryName,
			err,
		)
	}

	var repo Repository

	if registrationExists &&
		registration.RepositoryID != "" {

		repo = Repository{
			ID:   registration.RepositoryID,
			Name: registration.RepositoryName,
		}
	} else {
		var exists bool

		repo, exists, err = s.repositories.GetRepository(
			ctx,
			req.Project,
			req.RepositoryName,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"get repository %q: %w",
				req.RepositoryName,
				err,
			)
		}

		if !exists {
			repo, err = s.repositories.CreateRepository(
				ctx,
				req.Project,
				req.RepositoryName,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"create repository %q: %w",
					req.RepositoryName,
					err,
				)
			}
		}
	}

	branchExists, err := s.repositories.BranchExists(

		ctx,
		req.Project,
		repo.ID,
		defaultBranch,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"check branch %q in repository %q: %w",
			defaultBranch,
			req.RepositoryName,
			err,
		)
	}

	if !branchExists {

		err = s.repositories.PushFiles(
			ctx,
			req.Project,
			repo.ID,
			defaultBranch,
			files,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"push files to repository %q: %w",
				req.RepositoryName,
				err,
			)
		}

	}

	var pipeline Pipeline

	if registrationExists &&
		registration.BuildPipelineID != 0 {

		pipeline = Pipeline{
			ID:   registration.BuildPipelineID,
			Name: registration.BuildPipelineName,
		}
	} else {
		pipeline, err = s.pipelines.CreatePipeline(
			ctx,
			req.Project,
			repo.ID,
			buildPipelineName(req.RepositoryName),
			pipelineYamlPath,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"create build pipeline for repository %q: %w",
				req.RepositoryName,
				err,
			)
		}
	}

	return &CreateResult{
		Repository: RepositoryResult{
			ID:   repo.ID,
			Name: repo.Name,
			URL:  repo.WebURL,
		},
		Pipeline: PipelineResult{
			ID:   pipeline.ID,
			Name: pipeline.Name,
		},
	}, nil
}

func hasEmptyFilePath(
	files []templatepkg.TemplateFile,
) bool {
	for _, file := range files {
		if strings.TrimSpace(file.Path) == "" {
			return true
		}
	}

	return false
}

func findNonAbsolutePath(
	files []templatepkg.TemplateFile,
) (string, bool) {
	for _, file := range files {
		if !strings.HasPrefix(file.Path, "/") {
			return file.Path, true
		}
	}

	return "", false
}

func findInvalidPath(
	files []templatepkg.TemplateFile,
) (string, bool) {
	for _, file := range files {
		if file.Path == "/" {
			return file.Path, true
		}

		if path.Clean(file.Path) != file.Path {
			return file.Path, true
		}
	}

	return "", false
}

func findDuplicatePath(
	files []templatepkg.TemplateFile,
) (string, bool) {
	seen := make(map[string]struct{}, len(files))

	for _, file := range files {
		if _, exists := seen[file.Path]; exists {
			return file.Path, true
		}

		seen[file.Path] = struct{}{}
	}

	return "", false
}

func findFile(
	files []templatepkg.TemplateFile,
	filePath string,
) (templatepkg.TemplateFile, bool) {
	for _, file := range files {
		if file.Path == filePath {
			return file, true
		}
	}

	return templatepkg.TemplateFile{}, false
}
