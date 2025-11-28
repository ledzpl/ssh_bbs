package ui

import (
	"ag/internal/bbs"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) neonBanner(title, subtitle string) string {
	art := styleBannerArt.Render(`
   ___    ______  ____   ____ 
  / _ |  / __/ / / / /  / __/
 / __ | _\ \/ /_/ / /__/ _/  
/_/ |_|/___/\____/____/___/  
`)

	info := lipgloss.JoinVertical(
		lipgloss.Left,
		styleBannerTitle.Render(title),
		styleBannerMeta.Render(subtitle),
		styleDivider.Render(strings.Repeat("-", 34)),
	)

	return lipgloss.JoinHorizontal(lipgloss.Left, art, info)
}

func framedSection(title, content string) string {
	header := styleSectionTitle.Render("[" + title + "]")
	return lipgloss.JoinVertical(lipgloss.Left, header, stylePanel.Render(content))
}

func (m Model) accentBar() string {
	return styleDivider.Render(strings.Repeat("=-", fixedWidth/2))
}

func badgeLine(items ...string) string {
	if len(items) == 0 {
		return ""
	}

	badges := make([]string, 0, len(items))
	for _, item := range items {
		badges = append(badges, styleBadge.Render(item))
	}

	return strings.Join(badges, " ")
}

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

	return lipgloss.Place(fixedWidth, fixedHeight, lipgloss.Left, lipgloss.Top, s)
}

func (m Model) viewBoards() string {
	header := m.neonBanner("Signal Boards", fmt.Sprintf("user: %s • boards online: %d", m.username, len(m.boards)))
	s := header + "\n" + m.accentBar() + "\n\n"

	if len(m.boards) == 0 {
		empty := framedSection("Board Radar", styleDim.Render("No boards available."))
		s += empty + "\n"
		return s
	}

	var body strings.Builder
	body.WriteString(badgeLine(
		fmt.Sprintf("%d boards", len(m.boards)),
		"enter -> open",
		"q -> quit",
	))
	body.WriteString("\n\n")

	body.WriteString(fmt.Sprintf("%s  %s\n",
		styleTableHead.Width(boardNameColWidth).Render("Board Name"),
		styleTableHead.Width(boardPostsColWidth).Render("Posts"),
	))
	body.WriteString(styleDim.Render(strings.Repeat("-", boardTableWidth)))
	body.WriteString("\n")

	for i, b := range m.boards {
		style := styleTableRow
		if i == m.boardIdx {
			style = styleTableSelected
		}

		body.WriteString(fmt.Sprintf("%s  %s\n",
			style.Width(boardNameColWidth).Render(b.Name),
			style.Width(boardPostsColWidth).Render(fmt.Sprintf("%d posts", b.PostCount)),
		))
	}

	s += framedSection("Board Radar", body.String())
	s += "\n" + styleHelp.Render("j/k: navigate • enter: select • q: quit")
	return s
}

func (m Model) viewPosts() string {
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
	}

	totalPages := (len(displayPosts) + m.postsPerPage - 1) / m.postsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	header := m.neonBanner(
		"Board: "+m.activeBoard,
		fmt.Sprintf("%d post(s) • page %d/%d", len(displayPosts), m.page+1, totalPages),
	)

	s := header + "\n" + m.accentBar() + "\n\n"

	// Show search input if in search mode
	if m.searchMode {
		s += styleMetaLabel.Render("Search: ") + m.searchInput.View() + "\n\n"
	}

	if m.searchQuery != "" && !m.searchMode {
		s += styleCommentMeta.Render(fmt.Sprintf("Found %d result(s) for \"%s\"", len(displayPosts), m.searchQuery)) + "\n\n"
	}

	if len(displayPosts) == 0 {
		s += framedSection("Posts Stream", styleDim.Render("No posts found.")) + "\n"
		return s
	}

	// Pagination Logic
	start := m.page * m.postsPerPage
	end := start + m.postsPerPage
	if end > len(displayPosts) {
		end = len(displayPosts)
	}
	pagePosts := displayPosts[start:end]

	// If in viewPost state, show full screen detail
	if m.state == viewPost {
		p := m.activePost

		// Set viewport to fixed dimensions
		m.viewport.Height = fixedViewportHeight
		m.viewport.Width = fixedViewportWidth

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
		s += styleSectionTitle.Render("[Reading Signal]")
		s += "\n" + styleDetailBox.Render(detail)
		s += "\n" + styleHelp.Render("j/k: navigate • r: reply • c: comments • d: delete • b: back • q: quit")
		return s
	}

	// Normal view
	var table strings.Builder
	table.WriteString(badgeLine(
		fmt.Sprintf("page %d/%d", m.page+1, totalPages),
		fmt.Sprintf("%d posts", len(displayPosts)),
		"w -> write",
	))
	table.WriteString("\n\n")

	table.WriteString(fmt.Sprintf("  %s  %s  %s  %s\n",
		styleTableHead.Width(6).Render("ID"),
		styleTableHead.Width(40).Render("Title"),
		styleTableHead.Width(15).Render("Author"),
		styleTableHead.Width(20).Render("Date"),
	))
	table.WriteString(styleDim.Render(strings.Repeat("=", 85)))
	table.WriteString("\n")

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

		table.WriteString(fmt.Sprintf("%s %s %s %s %s\n",
			style.Render(indicator),
			style.Width(6).Render(fmt.Sprintf("%d", p.ID)),
			style.Width(40).Render(title),
			style.Width(15).Render(p.Author),
			style.Width(20).Render(p.CreatedAt.Format("06-01-02 15:04")),
		))
	}

	s += framedSection("Posts Stream", table.String())
	s += "\n" + styleHelp.Render(fmt.Sprintf("Page %d of %d • /: search • j/k: move • n/p: page • enter: read • w: write • b: back • q: quit", m.page+1, totalPages))
	return s
}

func (m Model) viewPost() string {
	// This function is now effectively unused when m.state == viewPost,
	// as viewPosts handles the split view.
	// However, if m.state were to be viewPost and viewPosts didn't handle it,
	// this would be the fallback. For now, it remains as it was, but its call
	// site in Model.View() has been changed.
	p := m.activePost

	// Set viewport to fixed height
	m.viewport.Height = fixedViewportHeight

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
		fixedWidth,
		fixedHeight-2, // Leave room for help text
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)

	// Add help text at the bottom
	help := "\n" + styleDim.Render("Esc/b: close • ↑↓: scroll • q: quit")

	return positioned + help
}

func (m Model) viewCompose() string {
	header := m.neonBanner("Compose", "Markdown supported • save with Ctrl+S")

	form := fmt.Sprintf("%s\n%s\n\n%s\n%s",
		styleMetaLabel.Render("Title:"),
		m.textInput.View(),
		styleMetaLabel.Render("Content (Markdown supported):"),
		m.textarea.View(),
	)

	body := framedSection("New Transmission", form)
	help := styleHelp.Render("Tab: switch fields • Ctrl+S: submit • Esc: cancel")

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		header,
		m.accentBar(),
		body,
		help,
	)
}

func (m Model) renderPostContent() string {
	content := m.activePost.Content

	// Try to render as Markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(fixedViewportWidth-4),
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
	header := m.neonBanner("Comments", m.activePost.Title)
	s := header + "\n" + m.accentBar() + "\n\n"

	if len(m.comments) == 0 {
		s += framedSection("Thread", styleDim.Render("No comments yet. Press 'r' to add one.")) + "\n"
		s += "\n" + styleHelp.Render("j/k: navigate • r: reply • b: back • q: quit")
		return s
	}

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
	m.viewport.Height = fixedViewportHeight
	m.viewport.Width = fixedViewportWidth
	m.viewport.SetContent(commentLines.String())

	// Build final view
	s += framedSection("Thread", m.viewport.View())
	s += "\n" + styleHelp.Render("j/k: navigate • r: reply • b: back • q: quit")
	return s
}
