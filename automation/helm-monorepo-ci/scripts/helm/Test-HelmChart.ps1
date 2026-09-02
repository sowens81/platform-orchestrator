[CmdletBinding()]
param (
    [Parameter(Mandatory)]
    [string]$ChartPath,

    [Parameter(Mandatory)]
    [string]$ChartName,

    [string]$KubernetesVersion = "1.35.0",

    [string]$OutputDirectory = "./artifacts"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

function Write-Step {
    param(
        [Parameter(Mandatory)]
        [string]$Message
    )

    Write-Host ""
    Write-Host "============================================================"
    Write-Host $Message
    Write-Host "============================================================"
}

function Invoke-Checked {
    param(
        [Parameter(Mandatory)]
        [scriptblock]$Command,

        [Parameter(Mandatory)]
        [string]$Description
    )

    Write-Step $Description
    & $Command

    if ($LASTEXITCODE -ne 0) {
        throw "$Description failed with exit code $LASTEXITCODE."
    }
}

$ChartPath = (Resolve-Path $ChartPath).Path
$OutputDirectory = [System.IO.Path]::GetFullPath($OutputDirectory)

$RenderedDirectory = Join-Path $OutputDirectory "rendered"
$TestDirectory = Join-Path $OutputDirectory "tests"
$PackageDirectory = Join-Path $OutputDirectory "packages"

New-Item -ItemType Directory -Force -Path $RenderedDirectory | Out-Null
New-Item -ItemType Directory -Force -Path $TestDirectory | Out-Null
New-Item -ItemType Directory -Force -Path $PackageDirectory | Out-Null

Write-Host "Chart Name         : $ChartName"
Write-Host "Chart Path         : $ChartPath"
Write-Host "Kubernetes Version : $KubernetesVersion"
Write-Host "Output Directory   : $OutputDirectory"

Write-Step "Validate chart structure"

foreach ($requiredFile in @("Chart.yaml", "values.yaml")) {
    $path = Join-Path $ChartPath $requiredFile

    if (-not (Test-Path $path)) {
        throw "Required Helm file not found: $path"
    }

    Write-Host "Found: $requiredFile"
}

Invoke-Checked -Description "Read Helm chart metadata" -Command {
    helm show chart $ChartPath
}

Invoke-Checked -Description "Build Helm dependencies" -Command {
    helm dependency build $ChartPath
}

Invoke-Checked -Description "Lint Helm chart" -Command {
    helm lint `
        $ChartPath `
        --strict `
        --kube-version $KubernetesVersion
}

$renderedManifest = Join-Path $RenderedDirectory "$ChartName.yaml"

Write-Step "Render Helm templates"

$rendered = & helm template `
    $ChartName `
    $ChartPath `
    --namespace default `
    --kube-version $KubernetesVersion

if ($LASTEXITCODE -ne 0) {
    throw "helm template failed with exit code $LASTEXITCODE."
}

[System.IO.File]::WriteAllLines(
    $renderedManifest,
    [string[]]$rendered,
    [System.Text.UTF8Encoding]::new($false)
)

Invoke-Checked -Description "Validate Kubernetes manifests with kubeconform" -Command {
    kubeconform `
        -strict `
        -summary `
        -kubernetes-version $KubernetesVersion `
        $renderedManifest
}

$unitTestResults = Join-Path $TestDirectory "helm-unittest.xml"

Invoke-Checked -Description "Run Helm unit tests" -Command {
    helm unittest `
        $ChartPath `
        --strict `
        --output-type JUnit `
        --output-file $unitTestResults
}

Write-Step "Validate generated schemas"

& "$PSScriptRoot/Test-HelmSchemas.ps1" `
    -ChartPath $ChartPath

Write-Step "Generate Helm documentation"

& helm-docs `
    --chart-search-root $ChartPath

if ($LASTEXITCODE -ne 0) {
    throw "helm-docs failed with exit code $LASTEXITCODE."
}

Write-Step "Check generated documentation"

& git diff --exit-code -- $ChartPath

if ($LASTEXITCODE -ne 0) {
    throw "Generated chart files differ from committed files. Regenerate schemas/docs and commit them."
}

Invoke-Checked -Description "Package Helm chart" -Command {
    helm package `
        $ChartPath `
        --destination $PackageDirectory
}

Write-Step "Helm CI completed successfully"
Write-Host "Chart '$ChartName' passed all validation."
