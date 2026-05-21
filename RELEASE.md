# Releases

## Versioning

Tags use [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH` (example: `v0.1.0`).

## Cut a release

1. Ensure `main` is green in [CI](https://github.com/venedicus/VAdlp/actions/workflows/ci.yml).
2. Update changelog notes if you keep them manually (optional; GitHub can generate release notes).
3. Tag and push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The [Release workflow](.github/workflows/release.yml) builds assets and publishes a GitHub Release.

## Assets

| Archive | Platform |
|---------|----------|
| `vadlp-linux-amd64.tar.gz` | Linux x86_64 |
| `vadlp-linux-arm64.tar.gz` | Linux ARM64 |
| `vadlp-windows-amd64.zip` | Windows x86_64 |
| `vadlp-darwin-arm64.tar.gz` | macOS Apple Silicon |
| `vadlp-darwin-amd64.tar.gz` | macOS Intel |

Each archive has a `.sha256` sidecar. `checksums.txt` lists all archives in one file.

## Verify download

```bash
sha256sum -c vadlp-linux-amd64.tar.gz.sha256
```

## Not in automated releases (yet)

- Windows/macOS code signing
- MSI, DMG, or AppImage installers
- Windows ARM64 builds

Track these as follow-ups when distribution requirements grow.
