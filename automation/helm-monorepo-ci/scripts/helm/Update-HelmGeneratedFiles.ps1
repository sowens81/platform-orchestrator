[CmdletBinding()]
param (
    [Parameter(Mandatory)]
    [string]$ChartPath
)

$ErrorActionPreference = "Stop"

$ChartPath = (Resolve-Path $ChartPath).Path

Write-Host "Generating schemas..."
& "$PSScriptRoot/New-HelmSchemas.ps1" `
    -ChartPath $ChartPath

Write-Host "Generating documentation..."
& helm-docs `
    --chart-search-root $ChartPath

if ($LASTEXITCODE -ne 0) {
    throw "helm-docs failed with exit code $LASTEXITCODE."
}

Write-Host ""
Write-Host "Generated files are current."
Write-Host "Review and commit:"
Write-Host "  $ChartPath/values.schema.json"
Write-Host "  $ChartPath/schema.json"
Write-Host "  $ChartPath/README.md"
