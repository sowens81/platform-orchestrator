package template

import (
	"fmt"
	"regexp"
	"strings"
)

var tokenPattern = regexp.MustCompile(`__[A-Z0-9_]+__`)

func replaceTokens(content string, values map[string]string) (string, error) {
	rendered := replaceKnownTokens(content, values)

	if token := findUnresolvedToken(rendered); token != "" {
		return "", fmt.Errorf("unresolved template token: %s", token)
	}

	return rendered, nil
}

func replaceKnownTokens(content string, values map[string]string) string {
	for key, value := range values {
		token := "__" + key + "__"
		content = strings.ReplaceAll(content, token, value)
	}

	return content
}

func findUnresolvedToken(content string) string {
	return tokenPattern.FindString(content)
}
