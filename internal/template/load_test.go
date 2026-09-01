package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndRenderFile_LoadsFileAndReplacesTokens(t *testing.T) {
	tempDir := t.TempDir()

	templatePath := filepath.Join(tempDir, "README.md")

	content := `
# __SERVICE_NAME__

Owner: __TEAM_NAME__
`

	err := os.WriteFile(
		templatePath,
		[]byte(content),
		0o644,
	)
	if err != nil {
		t.Fatalf("failed to create test template file: %v", err)
	}

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
		"TEAM_NAME":    "payments",
	}

	got, err := loadAndRenderFile(
		templatePath,
		values,
	)
	if err != nil {
		t.Fatalf("loadAndRenderFile() returned unexpected error: %v", err)
	}

	want := `
# payments-api

Owner: payments
`

	if got != want {
		t.Errorf("loadAndRenderFile() = %q, want %q", got, want)
	}
}

func TestLoadTemplateDirectory_LoadsFilesRecursively(t *testing.T) {
	tempDir := t.TempDir()

	nestedDir := filepath.Join(
		tempDir,
		"src",
		"__SERVICE_NAME__",
	)

	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	readmePath := filepath.Join(tempDir, "README.md")

	if err := os.WriteFile(
		readmePath,
		[]byte("# __SERVICE_NAME__"),
		0o644,
	); err != nil {
		t.Fatalf("failed to create README template: %v", err)
	}

	projectPath := filepath.Join(
		nestedDir,
		"__SERVICE_NAME__.csproj",
	)

	if err := os.WriteFile(
		projectPath,
		[]byte("<Project>__SERVICE_NAME__</Project>"),
		0o644,
	); err != nil {
		t.Fatalf("failed to create project template: %v", err)
	}

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
	}

	files, err := loadTemplateDirectory(tempDir, values)
	if err != nil {
		t.Fatalf(
			"loadTemplateDirectory() returned unexpected error: %v",
			err,
		)
	}

	if len(files) != 2 {
		t.Fatalf(
			"loadTemplateDirectory() returned %d files, want 2",
			len(files),
		)
	}

	got := map[string]string{}

	for _, file := range files {
		got[file.Path] = file.Content
	}

	want := map[string]string{
		"/README.md":                            "# payments-api",
		"/src/payments-api/payments-api.csproj": "<Project>payments-api</Project>",
	}

	for path, content := range want {
		if got[path] != content {
			t.Errorf(
				"file %q content = %q, want %q",
				path,
				got[path],
				content,
			)
		}
	}
}

func TestLoadTemplateDirectory_ReturnsErrorWhenDirectoryDoesNotExist(t *testing.T) {
	root := filepath.Join(
		t.TempDir(),
		"does-not-exist",
	)

	files, err := loadTemplateDirectory(
		root,
		map[string]string{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if files != nil {
		t.Errorf(
			"loadTemplateDirectory() returned files %v, want nil",
			files,
		)
	}
}

func TestLoadTemplateDirectory_ReturnsErrorWhenPathContainsUnknownToken(t *testing.T) {
	tempDir := t.TempDir()

	unknownTokenDir := filepath.Join(
		tempDir,
		"src",
		"__UNKNOWN_TOKEN__",
	)

	if err := os.MkdirAll(unknownTokenDir, 0o755); err != nil {
		t.Fatalf("failed to create template directory: %v", err)
	}

	templatePath := filepath.Join(
		unknownTokenDir,
		"Program.cs",
	)

	if err := os.WriteFile(
		templatePath,
		[]byte("namespace Test;"),
		0o644,
	); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	files, err := loadTemplateDirectory(
		tempDir,
		map[string]string{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if files != nil {
		t.Errorf(
			"loadTemplateDirectory() returned files %v, want nil",
			files,
		)
	}
}

func TestLoadTemplateDirectory_ReturnsErrorForDuplicateRenderedPaths(t *testing.T) {
	tempDir := t.TempDir()

	firstDir := filepath.Join(
		tempDir,
		"src",
		"__SERVICE_NAME__",
	)

	secondDir := filepath.Join(
		tempDir,
		"src",
		"payments-api",
	)

	if err := os.MkdirAll(firstDir, 0o755); err != nil {
		t.Fatalf("failed to create first template directory: %v", err)
	}

	if err := os.MkdirAll(secondDir, 0o755); err != nil {
		t.Fatalf("failed to create second template directory: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(firstDir, "config.yml"),
		[]byte("first"),
		0o644,
	); err != nil {
		t.Fatalf("failed to create first template file: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(secondDir, "config.yml"),
		[]byte("second"),
		0o644,
	); err != nil {
		t.Fatalf("failed to create second template file: %v", err)
	}

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
	}

	files, err := loadTemplateDirectory(
		tempDir,
		values,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if files != nil {
		t.Errorf(
			"loadTemplateDirectory() returned files %v, want nil",
			files,
		)
	}
}

func TestService_Load_LoadsNamedTemplate(
	t *testing.T,
) {
	templateRoot := t.TempDir()

	templatePath := filepath.Join(
		templateRoot,
		"go-api",
	)

	err := os.MkdirAll(
		templatePath,
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			templatePath,
			"template.yaml",
		),
		[]byte(`apiVersion: platform-orchestrator/v1
kind: RepositoryTemplate

metadata:
  name: go-api
  displayName: Go API
  description: Go API repository template
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME
      description: Name of the service
      example: payments-api

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	err = os.WriteFile(

		filepath.Join(
			templatePath,
			"README.md",
		),
		[]byte(
			"# __SERVICE_NAME__",
		),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		templateRoot,
	)

	rendered, err := service.Load(
		"go-api",
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

func TestService_Load_ExcludesTemplateMetadataFile(
	t *testing.T,
) {
	templateRoot := t.TempDir()

	templatePath := filepath.Join(
		templateRoot,
		"dotnet-api",
	)

	err := os.MkdirAll(
		templatePath,
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"create template directory: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			templatePath,
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

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
`),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"write template metadata: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			templatePath,
			"README.md",
		),
		[]byte(
			"# __SERVICE_NAME__",
		),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"write template file: %v",
			err,
		)
	}

	service := NewService(
		templateRoot,
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
		for _, file := range files {
			t.Logf(
				"loaded template file: path=%q content=%q",
				file.Path,
				file.Content,
			)
		}

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
