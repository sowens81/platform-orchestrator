package repository

const (
	buildPipelineSuffix   = "-build"
	releasePipelineSuffix = "-release"
)

func buildPipelineName(serviceName string) string {
	return serviceName + buildPipelineSuffix
}

func releasePipelineName(serviceName string) string {
	return serviceName + releasePipelineSuffix
}
