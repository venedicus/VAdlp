//go:build windows

package updater

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func lookPath(name string) (string, error) {
	return searchPath(name, windowsSearchPath())
}

func windowsSearchPath() string {
	seen := map[string]struct{}{}
	var parts []string
	add := func(s string) {
		for _, p := range filepath.SplitList(s) {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			key := strings.ToLower(p)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			parts = append(parts, p)
		}
	}
	add(os.Getenv("PATH"))
	if s, err := readEnvString(registry.CURRENT_USER, `Environment`, "Path"); err == nil {
		add(s)
	}
	if s, err := readEnvString(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "Path"); err == nil {
		add(s)
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

func readEnvString(key registry.Key, subkey, name string) (string, error) {
	k, err := registry.OpenKey(key, subkey, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	val, _, err := k.GetStringValue(name)
	return val, err
}

func searchPath(name, pathEnv string) (string, error) {
	if filepath.IsAbs(name) {
		if _, err := os.Stat(name); err == nil {
			return filepath.Clean(name), nil
		}
	}
	names := []string{name}
	if !strings.Contains(filepath.Base(name), ".") {
		names = append(names, name+".exe")
	}
	for _, dir := range filepath.SplitList(pathEnv) {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		for _, n := range names {
			p := filepath.Join(dir, n)
			if _, err := os.Stat(p); err == nil {
				return filepath.Clean(p), nil
			}
		}
	}
	return "", os.ErrNotExist
}
