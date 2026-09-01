package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestService_Load_ReturnsMetadataAndRenderedFiles(
	t *testing.T,
) {
	root := t.TempDir()

	templateRoot := filepath.Join(
		root,
		"dotnet-api",
	)

	err := os.MkdirAll(
		filepath.Join(
			templateRoot,
			".azuredevops",
		),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"MkdirAll() returned error: %v",
			err,
		)
	}

	err = os.WriteFile(
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
  pipelines:
    build:
      path: /.azuredevops/azure-pipelines-build.yml
    release:
      path: /.azuredevops/azure-pipelines-release.yml
`),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"WriteFile(template.yaml) returned error: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			templateRoot,
			"README.md",
		),
		[]byte("# __SERVICE_NAME__"),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"WriteFile(README.md) returned error: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			templateRoot,
			".azuredevops",
			"azure-pipelines-build.yml",
		),
		[]byte("trigger:\n- main\n"),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"WriteFile(build pipeline) returned error: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			templateRoot,
			".azuredevops",
			"azure-pipelines-release.yml",
		),
		[]byte("trigger: none\n"),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"WriteFile(release pipeline) returned error: %v",
			err,
		)
	}

	service := NewService(root)

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

	if rendered.Metadata.Metadata.Name != "dotnet-api" {
		t.Errorf(
			"metadata name = %q, want %q",
			rendered.Metadata.Metadata.Name,
			"dotnet-api",
		)
	}

	if rendered.Metadata.Spec.Pipelines.Build.Path !=
		"/.azuredevops/azure-pipelines-build.yml" {

		t.Errorf(
			"build pipeline path = %q",
			rendered.Metadata.Spec.Pipelines.Build.Path,
		)
	}

	if rendered.Metadata.Spec.Pipelines.Release.Path !=
		"/.azuredevops/azure-pipelines-release.yml" {

		t.Errorf(
			"release pipeline path = %q",
			rendered.Metadata.Spec.Pipelines.Release.Path,
		)
	}

	if len(rendered.Files) != 3 {
		t.Errorf(
			"file count = %d, want 3",
			len(rendered.Files),
		)
	}
}
