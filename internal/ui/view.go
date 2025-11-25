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
	case viewPosts, viewPost:
		s = m.viewPosts()
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

	// Build post list
	var postList string
	postList += fmt.Sprintf("%s %s\n",
		styleTableHead.Width(6).Render("ID"),
		styleTableHead.Width(22).Render("Title"),
	)

	for i, p := range pagePosts {
		style := styleTableRow
		currentIdx := start + i
		if currentIdx == m.postIdx {
			style = styleTableSelected
		}

		title := p.Title
		if len(title) > 20 {
			title = title[:17] + "..."
		}

		postList += fmt.Sprintf("%s %s\n",
			style.Width(6).Render(fmt.Sprintf("%d", p.ID)),
			style.Width(22).Render(title),
		)
	}

	postList += "\n" + styleDim.Render(fmt.Sprintf("Page %d/%d", m.page+1, totalPages))

	// If in viewPost state, show split pane
	if m.state == viewPost {
		// Adjust viewport for split pane
		// Account for: title (2 lines) + search (if active, 2 lines) + table header (1 line) + footer (2 lines) + padding
		headerLines := 5
		if m.searchMode || m.searchQuery != "" {
			headerLines += 2
		}

		maxContentHeight := m.height - headerLines - 10 // Extra padding for post title and meta
		if maxContentHeight < 5 {
			maxContentHeight = 5
		}
		m.viewport.Height = maxContentHeight
		m.viewport.Width = m.width - 35 // Reserve space for post list

		p := m.activePost

		// Post detail
		meta := fmt.Sprintf("%s %s\n%s %s",
			styleMetaLabel.Render("Author:"),
			styleMetaValue.Render(p.Author),
			styleMetaLabel.Render("Date:"),
			styleMetaValue.Render(p.CreatedAt.Format(time.RFC822)),
		)

		detail := fmt.Sprintf("%s\n\n%s\n\n%s",
			styleTitle.Render(p.Title),
			meta,
			m.viewport.View(),
		)

		detailBox := styleDetailBox.Render(detail)

		// Combine list and detail side by side
		listLines := strings.Split(postList, "\n")
		detailLines := strings.Split(detailBox, "\n")

		maxLines := len(listLines)
		if len(detailLines) > maxLines {
			maxLines = len(detailLines)
		}

		var combined string
		for i := 0; i < maxLines; i++ {
			var listLine, detailLine string
			if i < len(listLines) {
				listLine = listLines[i]
			}
			if i < len(detailLines) {
				detailLine = detailLines[i]
			}

			// Pad list to 30 chars
			listLine = lipgloss.NewStyle().Width(30).Render(listLine)
			combined += listLine + " " + detailLine + "\n"
		}

		s += combined
		s += styleDim.Render("j/k: move • Esc/b: close detail • q: quit")
		return s
	}

	// Normal view (no detail)
	s += fmt.Sprintf("%s %s %s %s\n",
		styleTableHead.Width(6).Render("ID"),
		styleTableHead.Width(40).Render("Title"),
		styleTableHead.Width(15).Render("Author"),
		styleTableHead.Width(20).Render("Date"),
	)

	for i, p := range pagePosts {
		style := styleTableRow
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

	s += "\n" + styleDim.Render(fmt.Sprintf("Page %d of %d • /: search • j/k: move • n/p: page • enter: read • w: write • b: back • q: quit", m.page+1, totalPages))
	return s
}

func (m Model) viewPost() string {
	// This function is now effectively unused when m.state == viewPost,
	// as viewPosts handles the split view.
	// However, if m.state were to be viewPost and viewPosts didn't handle it,
	// this would be the fallback. For now, it remains as it was, but its call
	// site in Model.View() has been changed.
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
