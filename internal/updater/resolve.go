package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"vadlp/internal/configdir"
)

type resolvedTool struct {
	Found   bool
	Path    string
	Version string
	Source  DepSource
}

func resolveTool(binName, versionFlag, customPath string) resolvedTool {
	customPath = strings.TrimSpace(customPath)
	if customPath != "" {
		for _, c := range buildCandidates(binName, customPath) {
			if c.source != SourceCustom {
				continue
			}
			if st := probeExact(c.path, versionFlag); st.Found {
				return resolvedTool{
					Found:   true,
					Path:    st.Path,
					Version: st.Version,
					Source:  c.source,
				}
			}
		}
		return resolvedTool{}
	}
	for _, c := range buildCandidates(binName, "") {
		if st := probeExact(c.path, versionFlag); st.Found {
			return resolvedTool{
				Found:   true,
				Path:    st.Path,
				Version: st.Version,
				Source:  c.source,
			}
		}
	}
	return resolvedTool{}
}

type candidate struct {
	path   string
	source DepSource
}

func buildCandidates(binName, customPath string) []candidate {
	seen := map[string]struct{}{}
	var out []candidate

	add := func(path string, source DepSource) {
		if path == "" {
			return
		}
		path = filepath.Clean(path)
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		out = append(out, candidate{path: path, source: source})
	}

	if customPath = strings.TrimSpace(customPath); customPath != "" {
		customPath = filepath.Clean(customPath)
		if st, err := os.Stat(customPath); err == nil {
			if st.IsDir() {
				add(filepath.Join(customPath, binName), SourceCustom)
			} else {
				add(customPath, SourceCustom)
			}
		} else {
			add(customPath, SourceCustom)
			add(filepath.Join(customPath, binName), SourceCustom)
		}
	}

	if toolsDir, err := configdir.ToolsDir(); err == nil {
		add(filepath.Join(toolsDir, binName), SourceManaged)
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		add(filepath.Join(exeDir, "bin", binName), SourceLocal)
		add(filepath.Join(exeDir, binName), SourceLocal)
		add(filepath.Join(exeDir, "..", "bin", binName), SourceLocal)
	}
	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, "bin", binName), SourceLocal)
	}

	if p, err := lookPath(binName); err == nil {
		add(p, SourceSystem)
	}
	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(binName), ".exe") {
		base := strings.TrimSuffix(binName, ".exe")
		if p, err := lookPath(base); err == nil {
			add(p, SourceSystem)
		}
	}
	if strings.EqualFold(binName, denoBinName()) {
		if home, err := os.UserHomeDir(); err == nil {
			add(filepath.Join(home, ".deno", "bin", binName), SourceSystem)
		}
	}

	return out
}

// ResolveYtDlpPath returns the yt-dlp binary path using unified search order.
func ResolveYtDlpPath(customPath string) (string, error) {
	st := resolveTool(ytDlpBinName(), "--version", customPath)
	if !st.Found {
		return "", fmt.Errorf("yt-dlp not found (install from https://github.com/yt-dlp/yt-dlp or use VAdlp's built-in installer)")
	}
	return st.Path, nil
}
