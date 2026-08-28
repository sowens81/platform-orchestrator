package azuredevops

import (
	"context"
	"testing"
)

func TestStaticTokenProvider_Token(t *testing.T) {
	provider := NewStaticTokenProvider(
		"test-token",
	)

	token, err := provider.Token(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"Token() returned unexpected error: %v",
			err,
		)
	}

	if token != "test-token" {
		t.Errorf(
			"token = %q, want %q",
			token,
			"test-token",
		)
	}
}

func TestStaticTokenProvider_ImplementsTokenProvider(t *testing.T) {
	var _ TokenProvider = NewStaticTokenProvider(
		"test-token",
	)
}
