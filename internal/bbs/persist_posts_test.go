package bbs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPostFileSaveLoad(t *testing.T) {
	dir := t.TempDir()
	store := PostFile{Dir: dir}
	board := "general"
	posts := []Post{
		{ID: 1, Title: "one", Content: "c1", Author: "a", CreatedAt: fixedNow()},
		{ID: 2, Title: "two", Content: "c2", Author: "b", CreatedAt: fixedNow().Add(time.Minute)},
	}
	if err := store.Save(board, posts); err != nil {
		t.Fatalf("Save: %v", err)
	}
	stat, err := os.Stat(filepath.Join(dir, board+".json"))
	if err != nil || stat.Size() == 0 {
		t.Fatalf("expected file written")
	}
	loaded, err := store.Load(board)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != len(posts) || loaded[1].Title != "two" || loaded[0].ID != 1 {
		t.Fatalf("unexpected loaded posts: %+v", loaded)
	}
}
