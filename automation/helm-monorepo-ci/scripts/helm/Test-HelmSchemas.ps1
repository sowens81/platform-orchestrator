[CmdletBinding()]
param (
    [Parameter(Mandatory)]
    [string]$ChartPath
)

$ErrorActionPreference = "Stop"

$ChartPath = (Resolve-Path $ChartPath).Path

$valuesSchemaFile = Join-Path $ChartPath "values.schema.json"
$schemaFile = Join-Path $ChartPath "schema.json"

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("helm-schema-" + [guid]::NewGuid())
$tempChart = Join-Path $tempRoot "chart"

try {
    New-Item -ItemType Directory -Force -Path $tempChart | Out-Null

    Copy-Item `
        -Path (Join-Path $ChartPath "*") `
        -Destination $tempChart `
        -Recurse `
        -Force

    & helm-schema `
        --chart-search-root $tempChart `
        --helm-docs-compatibility-mode

    if ($LASTEXITCODE -ne 0) {
        throw "helm-schema validation generation failed."
    }

    $generatedValuesSchema = Join-Path $tempChart "values.schema.json"

    if (-not (Test-Path $generatedValuesSchema)) {
        throw "Generated values.schema.json was not found."
    }

    if (-not (Test-Path $valuesSchemaFile)) {
        throw "Committed values.schema.json does not exist."
    }

    if (-not (Test-Path $schemaFile)) {
        throw "Committed schema.json does not exist."
    }

    $generated = Get-Content -Raw $generatedValuesSchema | ConvertFrom-Json | ConvertTo-Json -Depth 100
    $committedValues = Get-Content -Raw $valuesSchemaFile | ConvertFrom-Json | ConvertTo-Json -Depth 100
    $committedGeneric = Get-Content -Raw $schemaFile | ConvertFrom-Json | ConvertTo-Json -Depth 100

    if ($generated -ne $committedValues) {
        throw "values.schema.json is stale. Run New-HelmSchemas.ps1 and commit the result."
    }

    if ($generated -ne $committedGeneric) {
        throw "schema.json is stale. Run New-HelmSchemas.ps1 and commit the result."
    }

    Write-Host "values.schema.json is current."
    Write-Host "schema.json is current."
}
finally {
    if (Test-Path $tempRoot) {
        Remove-Item -Recurse -Force $tempRoot
    }
}
