# Build vadlp into bin/ (same output path as task build).
$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Force -Path bin | Out-Null

$version = "dev"
$commit = "none"
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
try {
    $version = (git describe --tags --always --dirty 2>$null)
    if (-not $version) { $version = "dev" }
    $commit = (git rev-parse --short HEAD 2>$null)
    if (-not $commit) { $commit = "none" }
} catch {
    # git optional
}

$ldflags = @(
    "-X", "vadlp/internal/version.Version=$version",
    "-X", "vadlp/internal/version.Commit=$commit",
    "-X", "vadlp/internal/version.BuildDate=$date"
)

go build -ldflags ($ldflags -join " ") -o bin/vadlp.exe ./cmd/vadlp
Write-Host "Built bin/vadlp.exe"
