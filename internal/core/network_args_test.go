package core

import "testing"

func TestAppendNetworkArgsCookiesAndDeno(t *testing.T) {
	cfg := DefaultConfig()
	cfg.UseCookiesFile = true
	cfg.CookiesFile = "/tmp/c.txt"
	cfg.Proxy = "http://127.0.0.1:8080"
	cfg.DenoPath = "/usr/bin/deno"
	args := AppendNetworkArgs(nil, cfg)
	if len(args) < 4 {
		t.Fatalf("args: %v", args)
	}
	if args[0] != "--cookies" || args[1] != "/tmp/c.txt" {
		t.Fatalf("cookies: %v", args)
	}
	if !containsArg(args, "--proxy") || !containsArg(args, "--js-runtimes") {
		t.Fatalf("missing network flags: %v", args)
	}
}

func TestDenoRuntimeArgsEmpty(t *testing.T) {
	if got := DenoRuntimeArgs("  "); got != nil {
		t.Fatalf("got %v", got)
	}
}
