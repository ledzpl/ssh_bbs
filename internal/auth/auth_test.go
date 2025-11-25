package auth

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigMissingOK(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Users) != 0 {
		t.Fatalf("expected empty users")
	}
}

func TestAuthenticateSuccess(t *testing.T) {
	a := NewAuthenticator(Config{Users: []User{{Username: "alice", Password: "pw"}}})
	in := bytes.NewBufferString("alice\npw\n")
	out := &bytes.Buffer{}

	user, ok := a.Authenticate(bufio.NewReader(in), out, "", 3)
	if !ok || user != "alice" {
		t.Fatalf("expected success, got %v %q", ok, user)
	}
}

func TestAuthenticateFail(t *testing.T) {
	a := NewAuthenticator(Config{Users: []User{{Username: "alice", Password: "pw"}}})
	in := bytes.NewBufferString("alice\nwrong\nagain\nbad\n")
	out := &bytes.Buffer{}

	_, ok := a.Authenticate(bufio.NewReader(in), out, "", 2)
	if ok {
		t.Fatalf("expected failure")
	}
}

func TestAuthenticateInitialUser(t *testing.T) {
	a := NewAuthenticator(Config{Users: []User{{Username: "bob", Password: "pw"}}})
	in := bytes.NewBufferString("pw\n")
	out := &bytes.Buffer{}
	user, ok := a.Authenticate(bufio.NewReader(in), out, "bob", 1)
	if !ok || user != "bob" {
		t.Fatalf("unexpected result %v %q", ok, user)
	}
}

func TestReadPasswordNoEcho(t *testing.T) {
	in := bytes.NewBufferString("secret\n")
	out := &bytes.Buffer{}
	got, err := readPassword(bufio.NewReader(in), out)
	if err != nil || got != "secret" {
		t.Fatalf("unexpected %q %v", got, err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no echo")
	}
}

func TestLoadConfigParses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, []byte(`{"users":[{"username":"a","password":"b"}]}`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Users) != 1 || cfg.Users[0].Username != "a" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
