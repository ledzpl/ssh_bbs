package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"ag/internal/bbs"
)

type sessionState int

const (
	viewBoards sessionState = iota
	viewPosts
	viewPost
	viewCompose
)

type Model struct {
	board    *bbs.BBS
	username string
	width    int
	height   int

	state       sessionState
	boards      []bbs.BoardSummary
	posts       []bbs.Post
	activeBoard string
	activePost  bbs.Post

	// Navigation
	boardIdx int
	postIdx  int

	// Pagination
	page         int
	postsPerPage int

	// Components
	viewport  viewport.Model
	textInput textinput.Model // For title
	textarea  textarea.Model  // For post content
	composing bool

	// Search
	searchMode  bool
	searchInput textinput.Model
	searchQuery string

	err error
}

func NewModel(board *bbs.BBS, username string) Model {
	ti := textinput.New()
	ti.Placeholder = "Title"
	ti.Focus()

	vp := viewport.New(0, 0)

	ta := textarea.New()
	ta.Placeholder = "Write your post content here..."
	ta.SetWidth(76)
	ta.SetHeight(10)

	si := textinput.New()
	si.Placeholder = "Search posts..."

	m := Model{
		board:        board,
		username:     username,
		state:        viewBoards,
		textInput:    ti,
		viewport:     vp,
		textarea:     ta,
		searchInput:  si,
		postsPerPage: 10,
	}
	m.refreshBoards()
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// --- Logic Helpers ---

func (m *Model) refreshBoards() {
	m.boards = m.board.ListBoards()
	if m.activeBoard == "" && len(m.boards) > 0 {
		m.activeBoard = m.boards[0].Name
	}
}

func (m *Model) refreshPosts() {
	posts, err := m.board.ListPosts(m.activeBoard)
	if err != nil {
		m.posts = nil
		return
	}
	m.posts = posts
	m.page = 0
}

func (m *Model) goBack() {
	switch m.state {
	case viewPosts:
		m.state = viewBoards
	case viewPost:
		m.state = viewPosts
	case viewCompose:
		m.state = viewPosts
		m.composing = false
	}
}

func (m *Model) startCompose() {
	m.state = viewCompose
	m.composing = true
	m.textInput.Reset()
	m.textInput.Focus()
	m.textarea.Reset()
	m.textarea.Blur()
}
