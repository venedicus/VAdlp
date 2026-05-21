# Security

## Supported versions

Security fixes are applied to the latest release and `main`.

| Version | Supported |
|---------|-----------|
| latest  | yes       |
| older   | no        |

## Reporting

Do not open public issues for exploitable vulnerabilities.

Email the maintainer via the address on their [GitHub profile](https://github.com/venedicus) or use [GitHub private vulnerability reporting](https://github.com/venedicus/VAdlp/security/advisories/new) if enabled.

Include steps to reproduce, impact, and your VAdlp version (Tools tab).

## Scope notes

VAdlp runs yt-dlp, ffmpeg, and deno as subprocesses and may download binaries from third-party URLs configured in `internal/updater`. Report issues in those tools to their upstream projects unless VAdlp mishandles paths, downloads, or execution.
