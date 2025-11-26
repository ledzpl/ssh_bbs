package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Retro Neon Colors
	colCyan      = lipgloss.Color("51")  // Bright Cyan
	colYellow    = lipgloss.Color("226") // Bright Yellow
	colGreen     = lipgloss.Color("46")  // Neon Green
	colDim       = lipgloss.Color("240") // Dark Gray
	colErr       = lipgloss.Color("196") // Bright Red
	colPurple    = lipgloss.Color("201") // Magenta/Pink
	colOrange    = lipgloss.Color("208") // Orange
	colBlue      = lipgloss.Color("33")  // Bright Blue
	colLightGray = lipgloss.Color("255") // White
	colBlack     = lipgloss.Color("16")  // Black

	// Styles
	styleTitle = lipgloss.NewStyle().
			Foreground(colCyan).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(colCyan)

	styleSelected = lipgloss.NewStyle().
			Foreground(colBlack).
			Background(colGreen).
			Bold(true)

	styleNormal = lipgloss.NewStyle().
			Foreground(colGreen)

	styleDim = lipgloss.NewStyle().
			Foreground(colDim)

	styleHeader = lipgloss.NewStyle().
			Foreground(colYellow).
			Bold(true).
			MarginBottom(1)

	// Table Styles
	styleTableHead = lipgloss.NewStyle().
			Foreground(colCyan).
			Bold(true).
			Padding(0, 1)

	styleTableRow = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(colGreen)

	styleTableSelected = lipgloss.NewStyle().
				Foreground(colBlack).
				Background(colGreen).
				Bold(true).
				Padding(0, 1)

	// Post Detail Styles - Enhanced
	styleDetailBox = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(colPurple).
			Padding(1, 2).
			MarginTop(1)

	styleMetaLabel = lipgloss.NewStyle().
			Foreground(colYellow).
			Bold(true)

	styleMetaValue = lipgloss.NewStyle().
			Foreground(colLightGray)

	// Comment Styles - New
	styleCommentSeparator = lipgloss.NewStyle().
				Foreground(colDim).
				Bold(true)

	styleCommentAuthor = lipgloss.NewStyle().
				Foreground(colBlue).
				Bold(true)

	styleCommentContent = lipgloss.NewStyle().
				Foreground(colLightGray)

	styleCommentMeta = lipgloss.NewStyle().
				Foreground(colDim).
				Italic(true)

	// Compose Styles
	styleComposeBox = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(colYellow).
			Padding(1, 2).
			MarginTop(1)

	styleHelp = lipgloss.NewStyle().
			Foreground(colDim).
			MarginTop(1)

	// Modal Styles
	styleModalOverlay = lipgloss.NewStyle().
				Foreground(colDim)

	styleModalDialog = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(colCyan).
				Padding(1, 2).
				Width(80)
)
