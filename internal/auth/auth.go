package auth

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// User is a simple username/password credential.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config describes the auth file format.
type Config struct {
	Users []User `json:"users"`
}

// Authenticator validates credentials loaded from a config file.
type Authenticator struct {
	users map[string]string
}

// LoadConfig reads users from a JSON file. Missing file yields empty config.
func LoadConfig(path string) (Config, error) {
	if path == "" {
		return Config{}, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read auth file: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse auth file: %w", err)
	}
	return cfg, nil
}

// NewAuthenticator builds an authenticator from config.
func NewAuthenticator(cfg Config) Authenticator {
	users := make(map[string]string, len(cfg.Users))
	for _, u := range cfg.Users {
		u.Username = strings.TrimSpace(u.Username)
		if u.Username == "" {
			continue
		}
		users[u.Username] = u.Password
	}
	return Authenticator{users: users}
}

// Enabled returns true when any users are configured.
func (a Authenticator) Enabled() bool {
	return len(a.users) > 0
}

// Authenticate prompts for username/password up to maxAttempts.
// io.Writer is used for prompts; bufio.Reader is reused for input.
func (a Authenticator) Authenticate(r *bufio.Reader, w io.Writer, initialUser string, maxAttempts int) (string, bool) {
	if !a.Enabled() {
		// No auth configured.
		if initialUser == "" {
			initialUser = "guest"
		}
		return initialUser, true
	}
	username := strings.TrimSpace(initialUser)
	if username == "" {
		fmt.Fprint(w, "Username: ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", false
		}
		username = strings.TrimSpace(line)
	}

	wantPass, ok := a.users[username]
	if !ok {
		fmt.Fprint(w, "Unknown user\r\n")
		return "", false
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		fmt.Fprint(w, "Password: ")
		pass, err := readPassword(r, w)
		if err != nil {
			return "", false
		}
		if pass == wantPass {
			fmt.Fprint(w, "\r\n")
			return username, true
		}
		fmt.Fprint(w, "\r\nInvalid password\r\n")
	}
	return "", false
}

// readPassword reads a line without echo.
func readPassword(r *bufio.Reader, w io.Writer) (string, error) {
	var buf []rune
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return "", err
		}
		switch ch {
		case '\r', '\n':
			return string(buf), nil
		case 0x7f, 0x08:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
		default:
			if ch < 0x20 && ch != '\t' {
				continue
			}
			buf = append(buf, ch)
		}
	}
}
