package template

import "testing"

func TestReplaceTokens_ReplacesKnownToken(t *testing.T) {
	input := "# __SERVICE_NAME__"

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
	}

	got, err := replaceTokens(input, values)
	if err != nil {
		t.Fatalf("replaceTokens() returned unexpected error: %v", err)
	}

	want := "# payments-api"

	if got != want {
		t.Errorf("replaceTokens() = %q, want %q", got, want)
	}
}

func TestReplaceTokens_ReplacesMultipleKnownTokens(t *testing.T) {
	input := `
# __SERVICE_NAME__

Owner: __TEAM_NAME__

Namespace: __NAMESPACE__
`

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
		"TEAM_NAME":    "payments",
		"NAMESPACE":    "payments-api",
	}

	got, err := replaceTokens(input, values)
	if err != nil {
		t.Fatalf("replaceTokens() returned unexpected error: %v", err)
	}

	want := `
# payments-api

Owner: payments

Namespace: payments-api
`

	if got != want {
		t.Errorf("replaceTokens() = %q, want %q", got, want)
	}
}

func TestReplaceTokens_ReplacesTokensInPath(t *testing.T) {
	input := "src/__SERVICE_NAME__/__SERVICE_NAME__.csproj"

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
	}

	got, err := replaceTokens(input, values)
	if err != nil {
		t.Fatalf("replaceTokens() returned unexpected error: %v", err)
	}

	want := "src/payments-api/payments-api.csproj"

	if got != want {
		t.Errorf("replaceTokens() = %q, want %q", got, want)
	}
}

func TestReplaceTokens_ReturnsErrorWhenUnknownTokenRemains(t *testing.T) {
	input := `
# __SERVICE_NAME__

Owner: __UNKNOWN_TOKEN__
`

	values := map[string]string{
		"SERVICE_NAME": "payments-api",
	}

	_, err := replaceTokens(input, values)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
