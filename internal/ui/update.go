package ui

import (
	"ag/internal/bbs"
	"fmt"
	"strings"

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

	// Calculate display posts for navigation bounds
	displayPosts := m.posts
	if m.searchQuery != "" {
		filtered := []bbs.Post{}
		query := strings.ToLower(m.searchQuery)
		for _, p := range m.posts {
			if strings.Contains(strings.ToLower(p.Title), query) ||
				strings.Contains(strings.ToLower(p.Content), query) {
				filtered = append(filtered, p)
			}
		}
		displayPosts = filtered
	}

	// Pagination bounds
	start := m.page * m.postsPerPage
	end := start + m.postsPerPage
	if end > len(displayPosts) {
		end = len(displayPosts)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			m.searchMode = true
			m.searchInput.Focus()
			return m, nil
		case "up", "k":
			if m.postIdx > start {
				m.postIdx--
			}
		case "down", "j":
			if m.postIdx < end-1 {
				m.postIdx++
			}
		case "n":
			totalPages := (len(displayPosts) + m.postsPerPage - 1) / m.postsPerPage
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
			if len(displayPosts) > 0 && m.postIdx < len(displayPosts) {
				m.activePost = displayPosts[m.postIdx]
				m.state = viewPost
				m.viewport.SetContent(m.renderPostContent())
				m.viewport.GotoTop()
			}
		case "w":
			m.state = viewCompose
			m.composing = true
			m.textInput.Focus()
			m.textInput.SetValue("")
			m.textarea.SetValue("")
		case "esc", "left", "h", "b":
			m.state = viewBoards
		case "q":
			return m, tea.Quit
		}
	}

	return m, cmd
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
		case "r":
			m.state = viewCompose
			m.composing = true
			m.commentMode = true
			m.textInput.SetValue("") // Not used for comments but good to clear
			m.textarea.SetValue("")
			m.textarea.Focus()
			return m, nil
		case "c":
			m.state = viewComments
			m.viewport.GotoTop()
			return m, nil
		case "d":
			// Delete post
			err := m.board.DeletePost(m.activeBoard, m.activePost.ID, m.username)
			if err != nil {
				m.err = err
			} else {
				m.refreshPosts()
				m.refreshBoards()
				m.state = viewPosts
			}
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
