package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type TemplateMetadata struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	Metadata TemplateMetadataDetails `yaml:"metadata"`
	Spec     TemplateSpec            `yaml:"spec"`
}

type TemplateMetadataDetails struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"displayName"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type TemplateSpec struct {
	RequiredValues []TemplateRequiredValue `yaml:"requiredValues"`
	Pipelines      TemplatePipelines       `yaml:"pipelines"`
}

type TemplateRequiredValue struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`
}

type TemplatePipelines struct {
	Build   TemplatePipeline `yaml:"build"`
	Release TemplatePipeline `yaml:"release"`
}

type TemplatePipeline struct {
	Path string `yaml:"path"`
}

func loadTemplateMetadata(
	templateRoot string,
) (*TemplateMetadata, error) {
	metadataPath := filepath.Join(
		templateRoot,
		"template.yaml",
	)

	content, err := os.ReadFile(
		metadataPath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"read template metadata: %w",
			err,
		)
	}

	var metadata TemplateMetadata

	if err := yaml.Unmarshal(
		content,
		&metadata,
	); err != nil {
		return nil, fmt.Errorf(
			"parse template metadata: %w",
			err,
		)
	}

	return &metadata, nil
}

func validateRequiredValues(
	metadata *TemplateMetadata,
	values map[string]string,
) error {
	for _, requiredValue := range metadata.Spec.RequiredValues {
		value, exists := values[requiredValue.Name]

		if !exists || strings.TrimSpace(value) == "" {
			return fmt.Errorf(
				"required template value %q is missing",
				requiredValue.Name,
			)
		}
	}

	return nil
}

func validateTemplateMetadata(
	name string,
	metadata *TemplateMetadata,
) error {
	if metadata.APIVersion != "platform-orchestrator/v1" {
		return fmt.Errorf(
			"unsupported template apiVersion %q",
			metadata.APIVersion,
		)
	}

	if metadata.Kind != "RepositoryTemplate" {
		return fmt.Errorf(
			"unsupported template kind %q",
			metadata.Kind,
		)
	}

	if metadata.Metadata.Name != name {
		return fmt.Errorf(
			"template metadata name %q does not match template %q",
			metadata.Metadata.Name,
			name,
		)
	}

	return nil
}
