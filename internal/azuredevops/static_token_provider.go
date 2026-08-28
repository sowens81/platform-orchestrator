package azuredevops

import "context"

type StaticTokenProvider struct {
	token string
}

func NewStaticTokenProvider(
	token string,
) *StaticTokenProvider {
	return &StaticTokenProvider{
		token: token,
	}
}

func (p *StaticTokenProvider) Token(
	ctx context.Context,
) (string, error) {
	return p.token, nil
}
