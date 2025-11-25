package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateComments(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b":
			m.state = viewPost
			return m, nil
		case "up", "k":
			if len(m.comments) > 0 {
				m.commentIdx = (m.commentIdx - 1 + len(m.comments)) % len(m.comments)
			}
		case "down", "j":
			if len(m.comments) > 0 {
				m.commentIdx = (m.commentIdx + 1) % len(m.comments)
			}
		case "r":
			// Reply - add comment to current post
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
	return m, nil
}
