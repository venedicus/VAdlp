package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Profile struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Config      Config `json:"config"`
}

var profileNameRe = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N} _.\-]{0,63}$`)

func ValidateProfileName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("empty profile name")
	}
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		return errors.New("invalid profile name")
	}
	if !profileNameRe.MatchString(name) {
		return errors.New("invalid profile name")
	}
	return nil
}

func profilesDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, err2 := os.UserHomeDir()
		if err2 != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	dir = filepath.Join(dir, "vadlp", "profiles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func profilePath(name string) (string, error) {
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func SaveProfile(p Profile) error {
	if err := ValidateProfileName(p.Name); err != nil {
		return err
	}
	p.Config = ConfigForProfile(p.Config)
	path, err := profilePath(p.Name)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func LoadProfile(name string) (Profile, error) {
	path, err := profilePath(name)
	if err != nil {
		return Profile{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	return p, json.Unmarshal(b, &p)
}

func ListProfiles() ([]string, error) {
	dir, err := profilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".json" {
			names = append(names, name[:len(name)-5])
		}
	}
	sort.Strings(names)
	return names, nil
}

func RenameProfile(oldName, newName string) error {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == newName {
		return nil
	}
	if err := ValidateProfileName(newName); err != nil {
		return err
	}
	p, err := LoadProfile(oldName)
	if err != nil {
		return err
	}
	p.Name = newName
	if err := SaveProfile(p); err != nil {
		return err
	}
	return DeleteProfile(oldName)
}

func DeleteProfile(name string) error {
	path, err := profilePath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
