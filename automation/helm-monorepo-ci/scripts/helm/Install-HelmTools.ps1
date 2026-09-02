[CmdletBinding()]
param (
    [string]$KubeconformVersion = "v0.7.0",
    [string]$HelmDocsVersion = "v1.14.2",
    [string]$HelmSchemaVersion = "latest",
    [string]$HelmUnitTestVersion = "1.1.2"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$binDirectory = if ($env:PIPELINE_WORKSPACE) {
    Join-Path $env:PIPELINE_WORKSPACE "bin"
}
else {
    Join-Path $HOME ".local/bin"
}

New-Item -ItemType Directory -Force -Path $binDirectory | Out-Null

$env:GOBIN = $binDirectory
$env:PATH = "${binDirectory}:$env:PATH"

function Invoke-Checked {
    param(
        [Parameter(Mandatory)]
        [scriptblock]$Command,

        [Parameter(Mandatory)]
        [string]$Description
    )

    Write-Host "==> $Description"
    & $Command

    if ($LASTEXITCODE -ne 0) {
        throw "$Description failed with exit code $LASTEXITCODE."
    }
}

Invoke-Checked -Description "Install kubeconform $KubeconformVersion" -Command {
    go install "github.com/yannh/kubeconform/cmd/kubeconform@$KubeconformVersion"
}

Invoke-Checked -Description "Install helm-docs $HelmDocsVersion" -Command {
    go install "github.com/norwoodj/helm-docs/cmd/helm-docs@$HelmDocsVersion"
}

Invoke-Checked -Description "Install helm-schema $HelmSchemaVersion" -Command {
    go install "github.com/dadav/helm-schema/cmd/helm-schema@$HelmSchemaVersion"
}

$installedPlugins = helm plugin list 2>$null | Out-String

if ($installedPlugins -notmatch "unittest") {
    Invoke-Checked -Description "Install helm-unittest $HelmUnitTestVersion" -Command {
        helm plugin install "oci://ghcr.io/helm-unittest/helm-unittest/unittest:$HelmUnitTestVersion"
    }
}

if ($env:TF_BUILD -eq "True") {
    Write-Host "##vso[task.prependpath]$binDirectory"
}

Write-Host ""
Write-Host "Installed tools:"
helm version
kubeconform -v
helm-docs --version
helm-schema --version
helm plugin list
