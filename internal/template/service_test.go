package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestService_Load_LoadsSelectedTemplate(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.MkdirAll(
		templateRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	if err := writeTemplateMetadata(
		templateRoot,
		"dotnet-api",
	); err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"README.md",
		),
		[]byte(
			"# __SERVICE_NAME__",
		),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	rendered, err := service.Load(
		"dotnet-api",
		map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	)
	if err != nil {
		t.Fatalf(
			"Load() returned unexpected error: %v",
			err,
		)
	}

	files := rendered.Files

	if len(files) != 1 {
		t.Fatalf(
			"file count = %d, want 1",
			len(files),
		)
	}

	if files[0].Path != "/README.md" {
		t.Errorf(
			"path = %q, want %q",
			files[0].Path,
			"/README.md",
		)
	}

	if files[0].Content != "# payments-api" {
		t.Errorf(
			"content = %q, want %q",
			files[0].Content,
			"# payments-api",
		)
	}
}

func TestService_Load_SelectsRequestedTemplate(
	t *testing.T,
) {
	root := t.TempDir()

	dotnetRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	reactRoot := filepath.Join(
		root,
		"react-app",
	)

	if err := os.MkdirAll(
		dotnetRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create dotnet template directory: %v",
			err,
		)
	}

	if err := os.MkdirAll(
		reactRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create react template directory: %v",
			err,
		)
	}

	if err := writeTemplateMetadata(
		dotnetRoot,
		"dotnet-api",
	); err != nil {
		t.Fatalf(
			"write dotnet template metadata: %v",
			err,
		)
	}

	if err := writeTemplateMetadata(
		reactRoot,
		"react-app",
	); err != nil {
		t.Fatalf(
			"write react template metadata: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			dotnetRoot,
			"README.md",
		),
		[]byte(
			"# .NET __SERVICE_NAME__",
		),
		0o644,
	); err != nil {
		t.Fatalf(
			"write dotnet template: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			reactRoot,
			"README.md",
		),
		[]byte(
			"# React __SERVICE_NAME__",
		),
		0o644,
	); err != nil {
		t.Fatalf(
			"write react template: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	rendered, err := service.Load(
		"react-app",
		map[string]string{
			"SERVICE_NAME": "customer-portal",
		},
	)
	if err != nil {
		t.Fatalf(
			"Load() returned unexpected error: %v",
			err,
		)
	}
	files := rendered.Files

	if len(files) != 1 {
		t.Fatalf(
			"file count = %d, want 1",
			len(files),
		)
	}

	if files[0].Content != "# React customer-portal" {
		t.Errorf(
			"content = %q, want %q",
			files[0].Content,
			"# React customer-portal",
		)
	}
}

func TestService_Load_ReturnsErrorWhenTemplateDoesNotExist(
	t *testing.T,
) {
	root := t.TempDir()

	service := NewService(
		root,
	)

	_, err := service.Load(
		"missing-template",
		map[string]string{},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	if !strings.Contains(
		err.Error(),
		`open template "missing-template"`,
	) {
		t.Errorf(
			"error = %q, want missing template context",
			err,
		)
	}
}

func TestService_Load_ReturnsErrorWhenTemplateIsNotDirectory(
	t *testing.T,
) {
	root := t.TempDir()

	templatePath := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.WriteFile(
		templatePath,
		[]byte("not a directory"),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	_, err := service.Load(
		"dotnet-api",
		map[string]string{},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	expected := `template "dotnet-api" is not a directory`

	if err.Error() != expected {
		t.Errorf(
			"error = %q, want %q",
			err,
			expected,
		)
	}
}

func TestService_Load_WrapsTemplateLoadingError(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.MkdirAll(
		templateRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	if err := writeTemplateMetadata(
		templateRoot,
		"dotnet-api",
	); err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"README.md",
		),
		[]byte(
			"# __MISSING_VALUE__",
		),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	_, err := service.Load(
		"dotnet-api",
		map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	if !strings.Contains(
		err.Error(),
		`load template "dotnet-api"`,
	) {
		t.Errorf(
			"error = %q, want template load context",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"unresolved template token: __MISSING_VALUE__",
	) {
		t.Errorf(
			"error = %q, want unresolved token error",
			err,
		)
	}
}

func TestService_Load_ReturnsErrorWhenRequiredTemplateValueIsMissing(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.MkdirAll(
		templateRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"template.yaml",
		),
		[]byte(`apiVersion: platform-orchestrator/v1
kind: RepositoryTemplate

metadata:
  name: dotnet-api
  displayName: .NET API
  description: Production-ready ASP.NET Core API service
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME
      description: Name of the service
      example: payments-api

    - name: NAMESPACE
      description: Root .NET namespace
      example: Company.Payments

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"README.md",
		),
		[]byte(
			"# __SERVICE_NAME__",
		),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	_, err := service.Load(
		"dotnet-api",
		map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	expected := `required template value "NAMESPACE" is missing`

	if err.Error() != expected {
		t.Errorf(
			"error = %q, want %q",
			err,
			expected,
		)
	}
}

func writeTemplateMetadata(
	templateRoot string,
	name string,
) error {
	content := `apiVersion: platform-orchestrator/v1
kind: RepositoryTemplate

metadata:
  name: ` + name + `
  displayName: Test Template
  description: Test repository template
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME
      description: Name of the service
      example: example-service

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`

	return os.WriteFile(
		filepath.Join(
			templateRoot,
			"template.yaml",
		),
		[]byte(content),
		0o644,
	)
}

func TestService_Load_ReturnsErrorWhenMetadataNameDoesNotMatchTemplate(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.MkdirAll(
		templateRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"template.yaml",
		),
		[]byte(`apiVersion: platform-orchestrator/v1
kind: RepositoryTemplate

metadata:
  name: react-app
  displayName: .NET API
  description: Production-ready ASP.NET Core API service
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"README.md",
		),
		[]byte(
			"# __SERVICE_NAME__",
		),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	_, err := service.Load(
		"dotnet-api",
		map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	expected := `template metadata name "react-app" does not match template "dotnet-api"`

	if err.Error() != expected {
		t.Errorf(
			"error = %q, want %q",
			err,
			expected,
		)
	}
}

func TestService_Load_ReturnsErrorWhenTemplateAPIVersionIsUnsupported(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.MkdirAll(
		templateRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"template.yaml",
		),
		[]byte(`apiVersion: platform-orchestrator/v2
kind: RepositoryTemplate

metadata:
  name: dotnet-api
  displayName: .NET API
  description: Production-ready ASP.NET Core API service
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	_, err := service.Load(
		"dotnet-api",
		map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	expected := `unsupported template apiVersion "platform-orchestrator/v2"`

	if err.Error() != expected {
		t.Errorf(
			"error = %q, want %q",
			err,
			expected,
		)
	}
}

func TestService_Load_ReturnsErrorWhenTemplateKindIsUnsupported(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	if err := os.MkdirAll(
		templateRoot,
		0o755,
	); err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			templateRoot,
			"template.yaml",
		),
		[]byte(`apiVersion: platform-orchestrator/v1
kind: ApplicationTemplate

metadata:
  name: dotnet-api
  displayName: .NET API
  description: Production-ready ASP.NET Core API service
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`),
		0o644,
	); err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	service := NewService(
		root,
	)

	_, err := service.Load(
		"dotnet-api",
		map[string]string{
			"SERVICE_NAME": "payments-api",
		},
	)

	if err == nil {
		t.Fatal(
			"Load() error = nil, want error",
		)
	}

	expected := `unsupported template kind "ApplicationTemplate"`

	if err.Error() != expected {
		t.Errorf(
			"error = %q, want %q",
			err,
			expected,
		)
	}
}
