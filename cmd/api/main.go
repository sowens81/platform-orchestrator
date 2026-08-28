package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/sowens81/platform-orchestrator/docs"

	"github.com/sowens81/platform-orchestrator/internal/api"
	"github.com/sowens81/platform-orchestrator/internal/azuredevops"
	"github.com/sowens81/platform-orchestrator/internal/repository"
	templatepkg "github.com/sowens81/platform-orchestrator/internal/template"
)

// @title Platform Orchestrator API
// @version 0.1.0
// @description API for provisioning application repositories and associated Azure DevOps resources.
// @BasePath /
func main() {
	azureDevOpsURL := requireEnvironmentVariable(
		"AZURE_DEVOPS_URL",
	)

	azureDevOpsToken := requireEnvironmentVariable(
		"AZURE_DEVOPS_TOKEN",
	)

	templateRoot := os.Getenv(
		"TEMPLATE_ROOT",
	)

	if templateRoot == "" {
		templateRoot = "./templates"
	}

	tokenProvider := azuredevops.NewStaticTokenProvider(
		azureDevOpsToken,
	)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	azureDevOpsClient := azuredevops.NewClient(
		azureDevOpsURL,
		httpClient,
		tokenProvider,
	)

	templateService := templatepkg.NewService(
		templateRoot,
	)

	repositoryService := repository.NewService(
		templateService,
		azureDevOpsClient,
		azureDevOpsClient,
	)

	repositoryHandler := api.NewRepositoryHandler(
		repositoryService,
	)

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	v1 := router.Group("/v1")
	{
		v1.POST(
			"/repositories",
			repositoryHandler.Create,
		)
	}

	router.GET(
		"/health",
		api.Health,
	)

	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(
			swaggerFiles.Handler,
		),
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf(
		"platform orchestrator listening on %s",
		server.Addr,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf(
			"HTTP server failed: %v",
			err,
		)
	}
}

func requireEnvironmentVariable(
	name string,
) string {
	value := os.Getenv(name)

	if value == "" {
		log.Fatalf(
			"required environment variable %s is not configured",
			name,
		)
	}

	return value
}
