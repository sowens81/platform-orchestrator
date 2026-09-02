# Helm Monorepo CI Example

This example provides a reusable Azure DevOps CI process for Helm charts stored at:

`src/<chart-name>/chart`

## CI checks

The pipeline performs:

- Helm dependency build
- Helm lint
- Helm template rendering
- Kubernetes resource validation with kubeconform
- Helm unit tests with helm-unittest
- Dynamic generation/validation of `values.schema.json`
- Dynamic generation/validation of `schema.json`
- Documentation generation with helm-docs
- Helm package creation

## Schema files

`values.schema.json` is the standard schema file consumed by Helm.

`schema.json` is generated as an identical copy of `values.schema.json`. This is useful for
downstream tooling that expects a generic `schema.json` filename while keeping a single source
of truth in `values.yaml` annotations.

## Generate files locally

```powershell
pwsh ./scripts/helm/Install-HelmTools.ps1

pwsh ./scripts/helm/Update-HelmGeneratedFiles.ps1 `
  -ChartPath ./src/payments-api/chart
```

This generates:

- `values.schema.json`
- `schema.json`
- `README.md`

## Run the full CI process locally

```powershell
pwsh ./scripts/helm/Test-HelmChart.ps1 `
  -ChartName payments-api `
  -ChartPath ./src/payments-api/chart `
  -KubernetesVersion 1.35.0
```

## Adding another chart

Copy the chart structure beneath:

`src/<new-chart-name>/chart`

and create a thin ADO pipeline that extends:

`.azuredevops/templates/helm-chart-ci-template.yml`
