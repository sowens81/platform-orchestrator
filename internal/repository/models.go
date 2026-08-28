package repository

type CreateRequest struct {
	Project        string            `json:"project"`
	RepositoryName string            `json:"repositoryName"`
	Template       string            `json:"template"`
	Values         map[string]string `json:"values"`
}

type Repository struct {
	ID     string
	Name   string
	WebURL string
}

type Pipeline struct {
	ID   int
	Name string
}

type CreateResult struct {
	Repository RepositoryResult `json:"repository"`
	Pipeline   PipelineResult   `json:"pipeline"`
}

type RepositoryResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PipelineResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
