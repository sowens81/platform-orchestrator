package repository

import "testing"

func TestBuildPipelineName(t *testing.T) {
	got := buildPipelineName("payments-api")
	want := "payments-api-build"

	if got != want {
		t.Errorf(
			"buildPipelineName() = %q, want %q",
			got,
			want,
		)
	}
}

func TestReleasePipelineName(t *testing.T) {
	got := releasePipelineName("payments-api")
	want := "payments-api-release"

	if got != want {
		t.Errorf(
			"releasePipelineName() = %q, want %q",
			got,
			want,
		)
	}
}
