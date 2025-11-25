package bbs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultBoards = []string{"general", "tech"}

// BoardListStore persists board names.
type BoardListStore interface {
	Save(names []string) error
}

// PostStore persists posts per board.
type PostStore interface {
	Load(board string) ([]Post, error)
	Save(board string, posts []Post) error
}

// BoardListLoader loads board names.
type BoardListLoader interface {
	Load() ([]string, error)
}

// BoardFile stores board names as JSON on disk.
// Format: {"boards":["general","tech"]}.
type BoardFile struct {
	Path string
}

func (f BoardFile) Load() ([]string, error) {
	if f.Path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read boards file: %w", err)
	}
	var wrapper struct {
		Boards []string `json:"boards"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse boards file: %w", err)
	}
	return normalizeBoardNames(wrapper.Boards), nil
}

func (f BoardFile) Save(names []string) error {
	if f.Path == "" {
		return nil
	}
	names = normalizeBoardNames(names)
	payload := struct {
		Boards []string `json:"boards"`
	}{Boards: names}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal boards: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return fmt.Errorf("make dir: %w", err)
	}
	tmp := f.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp boards file: %w", err)
	}
	if err := os.Rename(tmp, f.Path); err != nil {
		return fmt.Errorf("rename boards file: %w", err)
	}
	return nil
}

func normalizeBoardNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}
