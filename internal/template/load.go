package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type TemplateFile struct {
	Path    string
	Content string
}

func loadAndRenderFile(
	path string,
	values map[string]string,
) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return replaceTokens(string(content), values)
}

func loadTemplateDirectory(
	root string,
	values map[string]string,
) ([]TemplateFile, error) {
	var files []TemplateFile

	renderedPaths := make(map[string]struct{})

	err := filepath.WalkDir(
		root,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() {
				return nil
			}

			if entry.Name() == ".gitkeep" {
				return nil
			}

			relativePath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			if filepath.ToSlash(relativePath) == "template.yaml" {
				return nil
			}

			content, err := loadAndRenderFile(path, values)
			if err != nil {
				return err
			}

			renderedPath, err := replaceTokens(
				filepath.ToSlash(relativePath),
				values,
			)
			if err != nil {
				return err
			}

			renderedPath = "/" + renderedPath

			if _, exists := renderedPaths[renderedPath]; exists {
				return fmt.Errorf(
					"duplicate rendered template path: %s",
					renderedPath,
				)
			}

			renderedPaths[renderedPath] = struct{}{}

			files = append(
				files,
				TemplateFile{
					Path:    renderedPath,
					Content: content,
				},
			)

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return files, nil
}
