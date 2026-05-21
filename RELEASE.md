# Releases

## Versioning

Tags use [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH` (example: `v0.1.0`).

## Cut a release

1. Ensure `main` is green in [CI](https://github.com/venedicus/VAdlp/actions/workflows/ci.yml).
2. Update changelog notes if you keep them manually (optional; GitHub can generate release notes).
3. Tag and push:

```bash
git tag v0.1.1
git push origin v0.1.1
```

The [Release workflow](.github/workflows/release.yml) builds assets and publishes a GitHub Release.

## Assets

| File | Platform | Notes |
|------|----------|--------|
| `vadlp-linux-amd64.tar.gz` | Linux x86_64 | portable tarball |
| `vadlp-linux-amd64.AppImage` | Linux x86_64 | no install; needs FUSE on host |
| `vadlp-linux-arm64.tar.gz` | Linux ARM64 | portable tarball |
| `vadlp-windows-amd64.zip` | Windows x86_64 | portable zip |
| `vadlp-windows-arm64.zip` | Windows ARM64 | portable zip (from v0.1.1+) |
| `vadlp-darwin-arm64.tar.gz` | macOS Apple Silicon | portable tarball |
| `vadlp-darwin-arm64.dmg` | macOS Apple Silicon | disk image (from v0.1.1+) |
| `vadlp-darwin-amd64.tar.gz` | macOS Intel | portable tarball |
| `vadlp-darwin-amd64.dmg` | macOS Intel | disk image (from v0.1.1+) |

Each primary archive has a `.sha256` sidecar. `checksums.txt` on the release lists all payloads.

## Verify download

```bash
sha256sum -c vadlp-linux-amd64.tar.gz.sha256
```

## Code signing (optional, not enabled by default)

To sign release binaries in CI, add repository secrets and extend the workflow:

| Secret | Use |
|--------|-----|
| `APPLE_CERTIFICATE_BASE64` + `APPLE_CERTIFICATE_PASSWORD` | macOS `.dmg` / binary signing |
| `WINDOWS_CERTIFICATE_BASE64` + `WINDOWS_CERTIFICATE_PASSWORD` | Authenticode for `.exe` |

Without secrets, assets are unsigned (typical for open-source nightlies).

## Not in automated releases (yet)

- Windows **MSI** / macOS notarization pipeline
- Store packages (Microsoft Store, Homebrew cask)

## First release

`v0.1.0` ships portable tarballs/zip for five platforms. Installer-style assets (`.dmg`, `.AppImage`, Windows ARM64) ship from **v0.1.1** onward.
