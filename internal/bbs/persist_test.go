package bbs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBoardFileLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "boards.json")
	file := BoardFile{Path: path}

	names, err := file.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected empty load on missing file, got %v", names)
	}

	if err := file.Save([]string{"general", "tech", "tech", "  extra  "}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatalf("expected data written")
	}

	loaded, err := file.Load()
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if len(loaded) != 3 || loaded[0] != "general" || loaded[1] != "tech" || loaded[2] != "extra" {
		t.Fatalf("unexpected loaded names: %v", loaded)
	}
}
