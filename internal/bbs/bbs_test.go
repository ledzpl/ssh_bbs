package bbs

import (
	"testing"
	"time"
)

func fixedNow() time.Time {
	return time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
}

func TestAddAndGetPost(t *testing.T) {
	b := New(fixedNow)
	created, err := b.AddPost("general", "alice", "Hello", "world")
	if err != nil {
		t.Fatalf("AddPost: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected ID 1, got %d", created.ID)
	}
	if !created.CreatedAt.Equal(fixedNow()) {
		t.Fatalf("unexpected CreatedAt %v", created.CreatedAt)
	}

	found, err := b.GetPost("general", 1)
	if err != nil {
		t.Fatalf("GetPost: %v", err)
	}
	if found.Title != "Hello" || found.Content != "world" || found.Author != "alice" {
		t.Fatalf("unexpected post: %+v", found)
	}
}

func TestListBoardsAndPosts(t *testing.T) {
	b := New(fixedNow)
	_, _ = b.AddPost("general", "alice", "First", "post")
	_, _ = b.AddPost("tech", "bob", "Go", "rocks")

	boards := b.ListBoards()
	if len(boards) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(boards))
	}
	if boards[0].Name != "general" || boards[1].Name != "tech" {
		t.Fatalf("unexpected boards: %+v", boards)
	}
	if boards[0].PostCount != 1 || boards[1].PostCount != 1 {
		t.Fatalf("unexpected counts: %+v", boards)
	}

	posts, err := b.ListPosts("tech")
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if len(posts) != 1 || posts[0].Title != "Go" {
		t.Fatalf("unexpected posts: %+v", posts)
	}

	posts[0].Title = "mutated"
	again, _ := b.ListPosts("tech")
	if again[0].Title != "Go" {
		t.Fatalf("ListPosts returned shared slice")
	}
}

func TestErrors(t *testing.T) {
	b := New(fixedNow)
	if _, err := b.ListPosts("missing"); err != ErrBoardNotFound {
		t.Fatalf("expected ErrBoardNotFound, got %v", err)
	}
	if _, err := b.GetPost("missing", 1); err != ErrBoardNotFound {
		t.Fatalf("expected ErrBoardNotFound, got %v", err)
	}
	if _, err := b.GetPost("general", 99); err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
	if _, err := b.AddPost("general", "alice", "   ", ""); err != ErrEmptyTitle {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
}

type recordingStore struct {
	saves [][]string
	err   error
}

func (r *recordingStore) Save(names []string) error {
	r.saves = append(r.saves, append([]string(nil), names...))
	return r.err
}

func TestStoresBoardsOnCreate(t *testing.T) {
	store := &recordingStore{}
	b := NewWithBoards(fixedNow, []string{"general"}, store, nil)
	if _, err := b.AddPost("tech", "bob", "hello", "world"); err != nil {
		t.Fatalf("AddPost: %v", err)
	}
	if len(store.saves) == 0 {
		t.Fatalf("expected store save to be called")
	}
	last := store.saves[len(store.saves)-1]
	if len(last) != 2 || last[0] != "general" || last[1] != "tech" {
		t.Fatalf("unexpected stored names: %v", last)
	}
}
