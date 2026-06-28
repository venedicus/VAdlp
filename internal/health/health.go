package health

import (
	"context"
	"net/http"
	"time"

	"vadlp/internal/updater"
)

type Severity int

const (
	SeverityOK Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityCritical
)

type Issue struct {
	ID       string
	Severity Severity
	Key      string
	Params   map[string]interface{}
}

type Checker interface {
	Check() []Issue
}

type Monitor struct {
	checkers []Checker
}

func NewMonitor(checkers ...Checker) *Monitor {
	return &Monitor{checkers: checkers}
}

func (m *Monitor) CheckAll() []Issue {
	var out []Issue
	for _, c := range m.checkers {
		out = append(out, c.Check()...)
	}
	return out
}

func AttentionCount(issues []Issue) int {
	n := 0
	for _, iss := range issues {
		if iss.Severity >= SeverityWarning {
			n++
		}
	}
	return n
}

func WorstSeverity(issues []Issue) Severity {
	w := SeverityOK
	for _, iss := range issues {
		if iss.Severity > w {
			w = iss.Severity
		}
	}
	return w
}

// --- dependency checker ---

type DependencyChecker struct {
	Paths   func() updater.DependencyPaths
	Deps    func() []updater.DependencyInfo
}

func (d *DependencyChecker) Check() []Issue {
	if d.Deps == nil {
		return nil
	}
	var issues []Issue
	for _, dep := range d.Deps() {
		switch dep.Status {
		case updater.DepMissing:
			if dep.Level == updater.DepRequired {
				issues = append(issues, Issue{
					ID: "dep:" + string(dep.ID), Severity: SeverityCritical,
					Key: "health.dep.missing_required", Params: map[string]interface{}{"Name": string(dep.ID)},
				})
			} else if dep.Level == updater.DepRecommended {
				issues = append(issues, Issue{
					ID: "dep:" + string(dep.ID), Severity: SeverityWarning,
					Key: "health.dep.missing_recommended", Params: map[string]interface{}{"Name": string(dep.ID)},
				})
			}
		case updater.DepOutdated:
			issues = append(issues, Issue{
				ID: "dep:" + string(dep.ID), Severity: SeverityWarning,
				Key: "health.dep.outdated", Params: map[string]interface{}{
					"Name": dep.ID, "Version": dep.Version, "Latest": dep.LatestVer,
				},
			})
		case updater.DepError:
			issues = append(issues, Issue{
				ID: "dep:" + string(dep.ID), Severity: SeverityWarning,
				Key: "health.dep.error", Params: map[string]interface{}{"Name": string(dep.ID)},
			})
		}
	}
	return issues
}

// --- network checker ---

type NetworkChecker struct {
	URL string
}

func (n *NetworkChecker) Check() []Issue {
	url := n.URL
	if url == "" {
		url = "https://github.com"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return []Issue{{ID: "network", Severity: SeverityWarning, Key: "health.network.offline", Params: nil}}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Issue{{ID: "network", Severity: SeverityWarning, Key: "health.network.offline", Params: nil}}
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return []Issue{{ID: "network", Severity: SeverityWarning, Key: "health.network.offline", Params: nil}}
	}
	return nil
}
