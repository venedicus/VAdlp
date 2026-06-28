// Build metadata for local builds and CI (replaces build-metadata.sh).
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	version, commit, date := metadata()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--github-output":
			if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
				f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				defer f.Close()
				fmt.Fprintf(f, "version=%s\ncommit=%s\ndate=%s\n", version, commit, date)
			}
			return
		case "--field":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: buildmeta --field version|commit|date")
				os.Exit(2)
			}
			switch os.Args[2] {
			case "version":
				fmt.Println(version)
			case "commit":
				fmt.Println(commit)
			case "date":
				fmt.Println(date)
			default:
				fmt.Fprintln(os.Stderr, "unknown field:", os.Args[2])
				os.Exit(2)
			}
			return
		}
	}

	fmt.Printf("VERSION=%s\nCOMMIT=%s\nDATE=%s\n", version, commit, date)
}

func metadata() (version, commit, date string) {
	switch {
	case os.Getenv("GITHUB_REF_TYPE") == "tag":
		version = strings.TrimPrefix(os.Getenv("GITHUB_REF_NAME"), "v")
	case os.Getenv("GITHUB_REF_NAME") == "main", os.Getenv("GITHUB_REF_NAME") == "master":
		version = "main"
	default:
		if out, err := exec.Command("git", "describe", "--tags", "--always", "--dirty").Output(); err == nil {
			version = strings.TrimSpace(string(out))
		} else {
			version = "dev"
		}
	}

	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		commit = strings.TrimSpace(string(out))
	} else {
		commit = "none"
	}

	date = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	return version, commit, date
}
