package ui

import (
	"ag/internal/bbs"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
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
	case viewComments:
		s = m.viewComments()
	}

	if m.err != nil {
		s += "\n" + styleDim.Render("Error: "+m.err.Error())
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, s)
}

func (m Model) viewBoards() string {
	// ASCII Art Banner
	banner := lipgloss.NewStyle().
		Foreground(colCyan).
		Bold(true).
		Render(`
   _____ _____ _    _   ____  ____   _____ 
  / ____/ ____| |  | | |  _ \|  _ \ / ____|
 | (___| (___ | |__| | | |_) | |_) | (___  
  \___ \\___ \|  __  | |  _ <|  _ < \___ \ 
  ____) |___) | |  | | | |_) | |_) |____) |
 |_____/_____/|_|  |_| |____/|____/|_____/ 
`)

	s := banner + "\n\n"
	s += styleHeader.Render("Select a Board") + "\n\n"

	if len(m.boards) == 0 {
		s += styleDim.Render("No boards available.") + "\n"
		return s
	}

	// Enhanced table header
	s += fmt.Sprintf("%s  %s\n",
		styleTableHead.Width(30).Render("Board Name"),
		styleTableHead.Width(15).Render("Posts"),
	)
	s += styleDim.Render(strings.Repeat("-", 47)) + "\n"

	for i, b := range m.boards {
		style := styleTableRow
		if i == m.boardIdx {
			style = styleTableSelected
		}

		s += fmt.Sprintf("%s  %s\n",
			style.Width(30).Render(b.Name),
			style.Width(15).Render(fmt.Sprintf("%d posts", b.PostCount)),
		)
	}

	s += "\n" + styleDim.Render("j/k: navigate • enter: select • q: quit")
	return s
}

func (m Model) viewPosts() string {
	// Board header with style
	boardHeader := lipgloss.NewStyle().
		Foreground(colPurple).
		Bold(true).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colPurple).
		Render("Board: " + m.activeBoard)

	s := boardHeader + "\n\n"

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
			s += styleCommentMeta.Render(fmt.Sprintf("Found %d result(s) for \"%s\"", len(filtered), m.searchQuery)) + "\n\n"
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

	// If in viewPost state, show full screen detail
	if m.state == viewPost {
		p := m.activePost

		// Fixed box height - use most of the screen
		boxHeight := m.height - 6 // Leave room for header and help text
		if boxHeight < 10 {
			boxHeight = 10
		}

		// Set viewport to fill the box
		m.viewport.Height = boxHeight - 6 // Account for title, meta, borders
		m.viewport.Width = m.width - 10

		// Build ALL content (post + comments) for viewport
		var viewportContent strings.Builder

		// Post content
		viewportContent.WriteString(m.renderPostContent())

		// Add comments section
		if len(p.Comments) > 0 {
			viewportContent.WriteString("\n\n")
			viewportContent.WriteString(styleCommentSeparator.Render("--- Comments ---"))
			viewportContent.WriteString("\n\n")

			for i, c := range p.Comments {
				indent := ""
				prefix := "*"
				if c.ParentID > 0 {
					indent = "  "
					prefix = ">"
				}

				// Comment header
				commentHeader := fmt.Sprintf("%s%s %s",
					indent,
					prefix,
					styleCommentAuthor.Render(c.Author),
				)

				// Comment content
				commentBody := indent + "  " + styleCommentContent.Render(c.Content)

				viewportContent.WriteString(commentHeader + "\n")
				viewportContent.WriteString(commentBody + "\n")

				// Add spacing between comments
				if i < len(p.Comments)-1 {
					viewportContent.WriteString(styleDim.Render(indent+"  -----") + "\n")
				}
			}
		}

		// Set viewport content
		m.viewport.SetContent(viewportContent.String())

		// Post metadata
		commentCount := len(p.Comments)
		meta := fmt.Sprintf("%s %s | %s %s | %s %d",
			styleMetaLabel.Render("Author:"),
			styleMetaValue.Render(p.Author),
			styleMetaLabel.Render("Date:"),
			styleMetaValue.Render(p.CreatedAt.Format("2006-01-02 15:04")),
			styleMetaLabel.Render("Comments:"),
			commentCount,
		)

		// Build detail view
		detail := fmt.Sprintf("%s\n\n%s\n\n%s",
			styleTitle.Render(p.Title),
			meta,
			m.viewport.View(),
		)

		// Render with fixed height
		s += styleDetailBox.Render(detail)
		s += "\n" + styleDim.Render("j/k: navigate • r: reply • c: comments • d: delete • b: back • q: quit")
		return s
	}

	// Normal view
	s += fmt.Sprintf("  %s  %s  %s  %s\n",
		styleTableHead.Width(6).Render("ID"),
		styleTableHead.Width(40).Render("Title"),
		styleTableHead.Width(15).Render("Author"),
		styleTableHead.Width(20).Render("Date"),
	)
	s += styleDim.Render(strings.Repeat("=", 85)) + "\n"

	for i, p := range pagePosts {
		style := styleTableRow
		indicator := " "
		currentIdx := start + i
		if currentIdx == m.postIdx {
			style = styleTableSelected
			indicator = ">"
		}

		title := p.Title
		if len(title) > 35 {
			title = title[:35] + "..."
		}

		s += fmt.Sprintf("%s %s %s %s %s\n",
			style.Render(indicator),
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
	header := lipgloss.NewStyle().
		Foreground(colYellow).
		Bold(true).
		Padding(0, 1).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(colYellow).
		Render("Compose New Post")

	form := fmt.Sprintf("%s\n%s\n\n%s\n%s",
		styleMetaLabel.Render("Title:"),
		m.textInput.View(),
		styleMetaLabel.Render("Content (Markdown supported):"),
		m.textarea.View(),
	)

	help := styleDim.Render("Tab: switch fields • Ctrl+S: submit • Esc: cancel")

	return fmt.Sprintf("%s\n\n%s\n\n%s",
		header,
		styleComposeBox.Render(form),
		help,
	)
}

func (m Model) renderPostContent() string {
	content := m.activePost.Content

	// Try to render as Markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.viewport.Width-4),
	)

	if err == nil {
		rendered, err := renderer.Render(content)
		if err == nil {
			return rendered
		}
	}

	// Fallback to plain text if Markdown rendering fails
	return content
}

func (m Model) viewComments() string {
	// Header with style
	commentHeader := lipgloss.NewStyle().
		Foreground(colBlue).
		Bold(true).
		Padding(0, 1).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(colBlue).
		Render(fmt.Sprintf("Comments: %s", m.activePost.Title))

	var s string

	if len(m.comments) == 0 {
		s = commentHeader + "\n\n"
		s += styleDim.Render("No comments yet. Press 'r' to add one.") + "\n"
		s += "\n" + styleDim.Render("j/k: navigate • r: reply • b: back • q: quit")
		return s
	} else {
		// Build comment lines into a string
		var commentLines strings.Builder
		commentLines.WriteString(styleDim.Render(fmt.Sprintf("Total: %d comment(s)", len(m.comments))) + "\n\n")

		// Table header
		commentLines.WriteString(fmt.Sprintf("  %s  %s  %s\n",
			styleTableHead.Width(5).Render("#"),
			styleTableHead.Width(15).Render("Author"),
			styleTableHead.Width(53).Render("Content"),
		))
		commentLines.WriteString(styleDim.Render(strings.Repeat("=", 75)) + "\n")

		for i, c := range m.comments {
			style := styleTableRow
			indicator := " "
			if i == m.commentIdx {
				style = styleTableSelected
				indicator = ">"
			}

			// Determine prefix based on parent
			prefix := "*"
			indent := ""
			if c.ParentID > 0 {
				prefix = ">"
				indent = "  "
			}

			lines := strings.Split(c.Content, "\n")
			for li, line := range lines {
				line = strings.TrimRight(line, "\r")
				rendered := style.Width(50 - len(indent) - 2).Render(line)
				if len(rendered) > 50 {
					rendered = rendered[:47] + "..."
				}
				num := ""
				author := ""
				currIndicator := " "
				if li == 0 {
					num = fmt.Sprintf("%d", i+1)
					author = c.Author
					currIndicator = indicator
				}
				commentLines.WriteString(fmt.Sprintf("%s %s  %s  %s%s %s\n",
					style.Render(currIndicator),
					style.Width(5).Render(num),
					style.Width(15).Render(author),
					indent,
					prefix,
					rendered,
				))
			}
		}

		// Set viewport for comments
		m.viewport.Height = m.height - 6 // reserve header and help lines
		m.viewport.Width = m.width - 10
		m.viewport.SetContent(commentLines.String())

		// Build final view
		s = commentHeader + "\n\n" + m.viewport.View()
		s += "\n" + styleDim.Render("j/k: navigate • r: reply • b: back • q: quit")
		return s
	}
}
