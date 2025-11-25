package ui

import (
	"ag/internal/bbs"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var s string

	switch m.state {
	case viewBoards:
		s = m.viewBoards()
	case viewPosts:
		s = m.viewPosts()
	case viewPost:
		s = m.viewPost()
	case viewCompose:
		s = m.viewCompose()
	}

	if m.err != nil {
		s += "\n" + styleDim.Render("Error: "+m.err.Error())
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, s)
}

func (m Model) viewBoards() string {
	// ASCII Art Banner
	banner := `
 ███████╗███████╗██╗  ██╗    ██████╗ ██████╗ ███████╗
 ██╔════╝██╔════╝██║  ██║    ██╔══██╗██╔══██╗██╔════╝
 ███████╗███████╗███████║    ██████╔╝██████╔╝███████╗
 ╚════██║╚════██║██╔══██║    ██╔══██╗██╔══██╗╚════██║
 ███████║███████║██║  ██║    ██████╔╝██████╔╝███████║
 ╚══════╝╚══════╝╚═╝  ╚═╝    ╚═════╝ ╚═════╝ ╚══════╝`

	s := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Render(banner) + "\n"

	s += styleDim.Render(fmt.Sprintf("Welcome, %s!", m.username)) + "\n\n"

	// Header
	s += fmt.Sprintf("%s %s\n",
		styleTableHead.Width(30).Render("Name"),
		styleTableHead.Width(10).Render("Posts"),
	)

	// Rows
	for i, b := range m.boards {
		style := styleTableRow
		if i == m.boardIdx {
			style = styleTableSelected
		}
		s += fmt.Sprintf("%s %s\n",
			style.Width(30).Render(b.Name),
			style.Width(10).Render(fmt.Sprintf("%d", b.PostCount)),
		)
	}
	s += "\n" + styleDim.Render("j/k: move • enter: select • w: write • q: quit")
	return s
}

func (m Model) viewPosts() string {
	s := styleTitle.Render("Board: "+m.activeBoard) + "\n\n"

	// Show search input if in search mode
	if m.searchMode {
		s += styleMetaLabel.Render("Search: ") + m.searchInput.View() + "\n\n"
	}

	// Filter posts if search query exists
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
		if !m.searchMode {
			s += styleDim.Render(fmt.Sprintf("Filtered: %d results for \"%s\"", len(filtered), m.searchQuery)) + "\n\n"
		}
	}

	if len(displayPosts) == 0 {
		s += styleDim.Render("No posts found.") + "\n"
		return s
	}

	// Pagination Logic
	start := m.page * m.postsPerPage
	end := start + m.postsPerPage
	if end > len(displayPosts) {
		end = len(displayPosts)
	}
	pagePosts := displayPosts[start:end]
	totalPages := (len(displayPosts) + m.postsPerPage - 1) / m.postsPerPage

	// Header
	s += fmt.Sprintf("%s %s %s %s\n",
		styleTableHead.Width(6).Render("ID"),
		styleTableHead.Width(40).Render("Title"),
		styleTableHead.Width(15).Render("Author"),
		styleTableHead.Width(20).Render("Date"),
	)

	// Rows
	for i, p := range pagePosts {
		style := styleTableRow
		// Adjust index for selection highlighting
		currentIdx := start + i
		if currentIdx == m.postIdx {
			style = styleTableSelected
		}

		title := p.Title
		if len(title) > 38 {
			title = title[:35] + "..."
		}

		s += fmt.Sprintf("%s %s %s %s\n",
			style.Width(6).Render(fmt.Sprintf("%d", p.ID)),
			style.Width(40).Render(title),
			style.Width(15).Render(p.Author),
			style.Width(20).Render(p.CreatedAt.Format("06-01-02 15:04")),
		)
	}

	// Footer
	s += "\n" + styleDim.Render(fmt.Sprintf("Page %d of %d • /: search • j/k: move • n/p: page • enter: read • w: write • b: back • q: quit", m.page+1, totalPages))
	return s
}

func (m Model) viewPost() string {
	p := m.activePost

	// Adjust viewport for modal
	maxContentHeight := m.height - 15 // Reserve space for title, meta, borders, help
	if maxContentHeight < 5 {
		maxContentHeight = 5
	}
	m.viewport.Height = maxContentHeight

	// Metadata Header
	meta := fmt.Sprintf("%s %s • %s %s",
		styleMetaLabel.Render("Author:"),
		styleMetaValue.Render(p.Author),
		styleMetaLabel.Render("Date:"),
		styleMetaValue.Render(p.CreatedAt.Format(time.RFC822)),
	)

	// Modal content
	modalContent := fmt.Sprintf("%s\n\n%s\n\n%s",
		styleTitle.Render(p.Title),
		meta,
		m.viewport.View(),
	)

	dialog := styleModalDialog.Render(modalContent)

	// Center the dialog
	positioned := lipgloss.Place(
		m.width,
		m.height-2, // Leave room for help text
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)

	// Add help text at the bottom
	help := "\n" + styleDim.Render("Esc/b: close • ↑↓: scroll • q: quit")

	return positioned + help
}

func (m Model) viewCompose() string {
	header := styleTitle.Render("Compose Post")

	form := fmt.Sprintf("%s\n\n%s\n%s",
		m.textInput.View(),
		styleMetaLabel.Render("Content:"),
		m.textarea.View(),
	)

	help := styleHelp.Render("Tab: switch fields • Ctrl+S: submit • Esc: cancel")

	return fmt.Sprintf("%s\n%s\n%s",
		header,
		styleComposeBox.Render(form),
		help,
	)
}

func (m Model) renderPostContent() string {
	return m.activePost.Content
}
