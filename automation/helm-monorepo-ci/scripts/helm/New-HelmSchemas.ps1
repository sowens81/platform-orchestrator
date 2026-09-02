[CmdletBinding()]
param (
    [Parameter(Mandatory)]
    [string]$ChartPath
)

$ErrorActionPreference = "Stop"

$ChartPath = (Resolve-Path $ChartPath).Path

$valuesFile = Join-Path $ChartPath "values.yaml"
$valuesSchemaFile = Join-Path $ChartPath "values.schema.json"
$schemaFile = Join-Path $ChartPath "schema.json"

if (-not (Test-Path $valuesFile)) {
    throw "values.yaml was not found at '$valuesFile'."
}

Write-Host "Generating Helm values schema for: $ChartPath"

& helm-schema `
    --chart-search-root $ChartPath `
    --helm-docs-compatibility-mode

if ($LASTEXITCODE -ne 0) {
    throw "helm-schema failed with exit code $LASTEXITCODE."
}

if (-not (Test-Path $valuesSchemaFile)) {
    throw "helm-schema completed but '$valuesSchemaFile' was not created."
}

# Helm itself consumes values.schema.json.
# schema.json is intentionally maintained as an identical generic copy so that
# downstream tooling expecting schema.json can consume the same contract.
Copy-Item `
    -Path $valuesSchemaFile `
    -Destination $schemaFile `
    -Force

# Re-serialize both files consistently so diffs are deterministic.
foreach ($file in @($valuesSchemaFile, $schemaFile)) {
    $schema = Get-Content -Raw -Path $file | ConvertFrom-Json
    $json = $schema | ConvertTo-Json -Depth 100
    [System.IO.File]::WriteAllText(
        $file,
        ($json + [Environment]::NewLine),
        [System.Text.UTF8Encoding]::new($false)
    )
}

Write-Host "Generated:"
Write-Host "  $valuesSchemaFile"
Write-Host "  $schemaFile"
