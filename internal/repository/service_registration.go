package repository

import "context"

// ServiceRegistration records the platform resources associated with a service.
//
// Resource names provide deterministic human-readable conventions,
// while provider IDs are treated as the authoritative resource identities.
type ServiceRegistration struct {
	ServiceName string

	Project string

	RepositoryID   string
	RepositoryName string
	RepositoryURL  string

	BuildPipelineID   int
	BuildPipelineName string

	ReleasePipelineID   int
	ReleasePipelineName string

	TemplateName    string
	TemplateVersion string
}

// ServiceRegistrationStore persists and retrieves orchestration state for a
// provisioned service.
type ServiceRegistrationStore interface {
	Get(
		ctx context.Context,
		project string,
		serviceName string,
	) (ServiceRegistration, bool, error)

	Save(
		ctx context.Context,
		registration ServiceRegistration,
	) error
}
