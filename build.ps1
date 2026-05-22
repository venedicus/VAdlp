# Build vadlp into bin/ (same metadata as build.sh when bash is available).
$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Force -Path bin | Out-Null

$version = "dev"
$commit = "none"
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

if (Get-Command bash -ErrorAction SilentlyContinue) {
    bash scripts/build-metadata.sh | ForEach-Object {
        if ($_ -match '^([^=]+)=(.*)$') {
            switch ($Matches[1]) {
                "VERSION" { $version = $Matches[2] }
                "COMMIT" { $commit = $Matches[2] }
                "DATE" { $date = $Matches[2] }
            }
        }
    }
} else {
    try {
        $v = git describe --tags --always --dirty 2>$null
        if ($v) { $version = $v }
        $c = git rev-parse --short HEAD 2>$null
        if ($c) { $commit = $c }
    } catch { }
}

$ldflags = @(
    "-X", "vadlp/internal/version.Version=$version",
    "-X", "vadlp/internal/version.Commit=$commit",
    "-X", "vadlp/internal/version.BuildDate=$date"
)

$env:CGO_ENABLED = "1"
go build -ldflags ($ldflags -join " ") -o bin/vadlp.exe ./cmd/vadlp
Write-Host "Built bin/vadlp.exe ($version)"
