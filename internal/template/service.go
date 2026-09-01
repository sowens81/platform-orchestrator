package template

import (
	"fmt"
	"os"
	"path/filepath"
)

type Service struct {
	root string
}

func NewService(
	root string,
) *Service {
	return &Service{
		root: root,
	}
}

func (s *Service) Load(
	name string,
	values map[string]string,
) (RenderedTemplate, error) {
	templateRoot := filepath.Join(
		s.root,
		name,
	)

	info, err := os.Stat(
		templateRoot,
	)
	if err != nil {
		return RenderedTemplate{}, fmt.Errorf(
			"open template %q: %w",
			name,
			err,
		)
	}

	if !info.IsDir() {
		return RenderedTemplate{}, fmt.Errorf(
			"template %q is not a directory",
			name,
		)
	}

	metadata, err := loadTemplateMetadata(
		templateRoot,
	)
	if err != nil {
		return RenderedTemplate{}, fmt.Errorf(
			"load template %q metadata: %w",
			name,
			err,
		)
	}

	if err := validateTemplateMetadata(
		name,
		metadata,
	); err != nil {
		return RenderedTemplate{}, err
	}

	if err := validateRequiredValues(
		metadata,
		values,
	); err != nil {
		return RenderedTemplate{}, err
	}

	files, err := loadTemplateDirectory(
		templateRoot,
		values,
	)
	if err != nil {
		return RenderedTemplate{}, fmt.Errorf(
			"load template %q: %w",
			name,
			err,
		)
	}

	return RenderedTemplate{
		Metadata: *metadata,
		Files:    files,
	}, nil
}
