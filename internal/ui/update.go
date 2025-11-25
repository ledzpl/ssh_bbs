package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5 // Leave room for header/footer
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Handle global keys if not composing or if composing but specific keys
	if !m.composing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "b", "left":
				m.goBack()
				return m, nil
			case "w":
				if m.state == viewBoards || m.state == viewPosts {
					m.startCompose()
					return m, nil
				}
			}
		}
	}

	switch m.state {
	case viewBoards:
		m, cmd = m.updateBoards(msg)
		cmds = append(cmds, cmd)
	case viewPosts:
		m, cmd = m.updatePosts(msg)
		cmds = append(cmds, cmd)
	case viewPost:
		m, cmd = m.updatePostView(msg)
		cmds = append(cmds, cmd)
	case viewCompose:
		m, cmd = m.updateCompose(msg)
		cmds = append(cmds, cmd)
	case viewComments:
		m, cmd = m.updateComments(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateBoards(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if len(m.boards) > 0 {
				m.boardIdx = (m.boardIdx - 1 + len(m.boards)) % len(m.boards)
			}
		case "down", "j":
			if len(m.boards) > 0 {
				m.boardIdx = (m.boardIdx + 1) % len(m.boards)
			}
		case "enter", "right", "l":
			if len(m.boards) > 0 {
				m.activeBoard = m.boards[m.boardIdx].Name
				m.refreshPosts()
				// Reset search when entering a new board
				m.searchMode = false
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.state = viewPosts
				m.postIdx = 0
			}
		}
	}
	return m, nil
}

func (m Model) updatePosts(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle search mode
	if m.searchMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.searchInput.SetValue("")
				return m, nil
			case "enter":
				m.searchMode = false
				return m, nil
			}
		}
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchQuery = m.searchInput.Value()
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			m.searchMode = true
			m.searchInput.Focus()
			return m, nil
		case "up", "k":
			if len(m.posts) > 0 {
				m.postIdx = (m.postIdx - 1 + len(m.posts)) % len(m.posts)
			}
		case "down", "j":
			if len(m.posts) > 0 {
				m.postIdx = (m.postIdx + 1) % len(m.posts)
			}
		case "n":
			totalPages := (len(m.posts) + m.postsPerPage - 1) / m.postsPerPage
			if m.page < totalPages-1 {
				m.page++
				m.postIdx = m.page * m.postsPerPage
			}
		case "p":
			if m.page > 0 {
				m.page--
				m.postIdx = m.page * m.postsPerPage
			}
		case "enter", "right", "l":
			if len(m.posts) > 0 {
				m.activePost = m.posts[m.postIdx]
				m.state = viewPost
				m.viewport.SetContent(m.renderPostContent())
				m.viewport.GotoTop()
			}
		}
	}
	return m, nil
}

func (m Model) updatePostView(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle modal close keys
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b", "left":
			m.state = viewPosts
			return m, nil
		case "c":
			// View comments (legacy - now shown inline)
			comments, err := m.board.ListComments(m.activeBoard, m.activePost.ID)
			if err == nil {
				m.comments = comments
				m.commentIdx = 0
				m.state = viewComments
			}
			return m, nil
		case "r":
			// Reply directly
			m.commentMode = true
			m.state = viewCompose
			m.composing = true
			m.textInput.Reset()
			m.textInput.SetValue("Comment")
			m.textInput.Blur()
			m.textarea.Reset()
			m.textarea.Focus()
			return m, nil
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) updateCompose(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Toggle between title and content
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.textarea.Focus()
			} else {
				m.textarea.Blur()
				m.textInput.Focus()
			}
			return m, nil
		case "ctrl+s":
			// Submit with Ctrl+S
			title := m.textInput.Value()
			content := m.textarea.Value()
			if content == "" {
				m.err = fmt.Errorf("content cannot be empty")
				return m, nil
			}

			if m.commentMode {
				// Add comment
				_, err := m.board.AddComment(m.activeBoard, m.activePost.ID, m.username, content, 0)
				if err != nil {
					m.err = err
				} else {
					// Refresh post to show new comment
					m.refreshPosts()
					// Find and update activePost
					for _, p := range m.posts {
						if p.ID == m.activePost.ID {
							m.activePost = p
							break
						}
					}
					m.state = viewPost
					m.composing = false
					m.commentMode = false
				}
			} else {
				// Add post
				if title == "" {
					m.err = fmt.Errorf("title cannot be empty")
					return m, nil
				}
				_, err := m.board.AddPost(m.activeBoard, m.username, title, content)
				if err != nil {
					m.err = err
				} else {
					m.refreshPosts()
					m.refreshBoards()
					m.state = viewPosts
					m.composing = false
				}
			}
			return m, nil
		case "esc":
			m.composing = false
			// Return to appropriate view
			if m.commentMode {
				m.commentMode = false
				m.state = viewPost
			} else {
				m.state = viewPosts
			}
			return m, nil
		}
	}

	// Update the focused component
	if m.textInput.Focused() {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.textarea, cmd = m.textarea.Update(msg)
	}

	return m, cmd
}
