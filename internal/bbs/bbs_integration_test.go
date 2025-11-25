package bbs

import "testing"

// Integration-like test to ensure posts reload from store.
func TestLoadsPostsFromStore(t *testing.T) {
	postStore := &memoryPostStore{data: map[string][]Post{
		"general": {{ID: 2, Title: "old", Content: "c", Author: "a"}},
	}}
	b := NewWithBoards(fixedNow, []string{"general"}, nil, postStore)
	posts, err := b.ListPosts("general")
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if len(posts) != 1 || posts[0].ID != 2 {
		t.Fatalf("unexpected posts: %+v", posts)
	}
	if _, err := b.AddPost("general", "alice", "new", "c2"); err != nil {
		t.Fatalf("AddPost: %v", err)
	}
	if !postStore.saved {
		t.Fatalf("expected post store Save to be called")
	}
}

type memoryPostStore struct {
	data  map[string][]Post
	saved bool
}

func (m *memoryPostStore) Load(board string) ([]Post, error) {
	return m.data[board], nil
}

func (m *memoryPostStore) Save(board string, posts []Post) error {
	m.saved = true
	m.data[board] = append([]Post(nil), posts...)
	return nil
}
