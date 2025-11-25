package bbs

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Post represents a single message on a board.
type Post struct {
	ID        int
	Title     string
	Content   string
	Author    string
	CreatedAt time.Time
}

// Board keeps ordered posts.
type Board struct {
	Name   string
	mu     sync.RWMutex
	posts  []Post
	nextID int
}

// BBS stores boards and posts in memory.
type BBS struct {
	mu     sync.RWMutex
	boards map[string]*Board
	order  []string
	now    func() time.Time
	store  BoardListStore
	posts  PostStore
}

var (
	// ErrBoardNotFound signals an unknown board.
	ErrBoardNotFound = errors.New("board not found")
	// ErrPostNotFound signals an unknown post.
	ErrPostNotFound = errors.New("post not found")
	// ErrEmptyTitle signals a missing post title.
	ErrEmptyTitle = errors.New("title is required")
)

// New returns an in-memory BBS with a default "general" board.
func New(now func() time.Time) *BBS {
	return NewWithBoards(now, nil, nil, nil)
}

// NewWithBoards builds a BBS with provided board names and optional stores.
// If names is empty, defaults are used. Board stores are invoked when new boards are created.
// Post stores persist per-board posts.
func NewWithBoards(now func() time.Time, names []string, store BoardListStore, posts PostStore) *BBS {
	if now == nil {
		now = time.Now
	}
	names = normalizeBoardNames(names)
	if len(names) == 0 {
		names = defaultBoards
	}
	boards := make(map[string]*Board, len(names))
	for _, name := range names {
		boards[name] = &Board{Name: name, nextID: 1}
	}
	b := &BBS{
		boards: boards,
		order:  append([]string(nil), names...),
		now:    now,
		store:  store,
		posts:  posts,
	}
	b.loadPosts()
	return b
}

// BoardSummary describes a board and its stats.
type BoardSummary struct {
	Name      string
	PostCount int
}

// ListBoards returns board summaries sorted by name.
func (b *BBS) ListBoards() []BoardSummary {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]BoardSummary, 0, len(b.order))
	for _, name := range b.order {
		board := b.boards[name]
		board.mu.RLock()
		count := len(board.posts)
		board.mu.RUnlock()
		out = append(out, BoardSummary{
			Name:      name,
			PostCount: count,
		})
	}
	return out
}

// ListPosts returns copies of posts for a board ordered by ID.
func (b *BBS) ListPosts(boardName string) ([]Post, error) {
	board, ok := b.board(boardName)
	if !ok {
		return nil, ErrBoardNotFound
	}
	board.mu.RLock()
	defer board.mu.RUnlock()

	posts := make([]Post, len(board.posts))
	copy(posts, board.posts)
	return posts, nil
}

// AddPost adds a post to the board, creating the board implicitly.
func (b *BBS) AddPost(boardName, author, title, content string) (Post, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Post{}, ErrEmptyTitle
	}
	if author == "" {
		author = "anonymous"
	}
	board, err := b.ensureBoard(boardName)
	if err != nil {
		return Post{}, fmt.Errorf("ensure board: %w", err)
	}

	board.mu.Lock()
	defer board.mu.Unlock()

	post := Post{
		ID:        board.nextID,
		Title:     title,
		Content:   strings.TrimSpace(content),
		Author:    author,
		CreatedAt: b.now(),
	}
	board.nextID++
	board.posts = append(board.posts, post)

	if b.posts != nil {
		if err := b.posts.Save(board.Name, board.posts); err != nil {
			return Post{}, fmt.Errorf("store post: %w", err)
		}
	}
	return post, nil
}

// GetPost returns one post by ID.
func (b *BBS) GetPost(boardName string, id int) (Post, error) {
	board, ok := b.board(boardName)
	if !ok {
		return Post{}, ErrBoardNotFound
	}
	board.mu.RLock()
	defer board.mu.RUnlock()
	for _, p := range board.posts {
		if p.ID == id {
			return p, nil
		}
	}
	return Post{}, ErrPostNotFound
}

func (b *BBS) board(name string) (*Board, bool) {
	b.mu.RLock()
	board, ok := b.boards[name]
	b.mu.RUnlock()
	return board, ok
}

func (b *BBS) ensureBoard(name string) (*Board, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "general"
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if board, ok := b.boards[name]; ok {
		return board, nil
	}
	board := &Board{Name: name, nextID: 1}
	b.boards[name] = board
	b.order = append(b.order, name)
	if b.store != nil {
		if err := b.store.Save(boardNames(b.boards, b.order)); err != nil {
			return nil, err
		}
	}
	return board, nil
}

func (b *BBS) loadPosts() {
	if b.posts == nil {
		return
	}
	for _, name := range b.order {
		posts, err := b.posts.Load(name)
		if err != nil {
			continue
		}
		board := b.boards[name]
		for _, p := range posts {
			board.posts = append(board.posts, p)
			if p.ID >= board.nextID {
				board.nextID = p.ID + 1
			}
		}
	}
}

func boardNames(m map[string]*Board, order []string) []string {
	names := make([]string, 0, len(m))
	for _, name := range order {
		if _, ok := m[name]; ok {
			names = append(names, name)
		}
	}
	return names
}
