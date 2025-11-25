package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Enhanced Colors
	colCyan      = lipgloss.Color("86")
	colYellow    = lipgloss.Color("220")
	colGreen     = lipgloss.Color("76")
	colDim       = lipgloss.Color("240")
	colErr       = lipgloss.Color("196")
	colPurple    = lipgloss.Color("141")
	colOrange    = lipgloss.Color("208")
	colBlue      = lipgloss.Color("39")
	colLightGray = lipgloss.Color("250")

	// Styles
	styleTitle = lipgloss.NewStyle().
			Foreground(colCyan).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(colCyan)

	styleSelected = lipgloss.NewStyle().
			Foreground(colGreen).
			Bold(true)

	styleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

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
			Padding(0, 1)

	styleTableSelected = lipgloss.NewStyle().
				Foreground(colGreen).
				Bold(true).
				Background(lipgloss.Color("235")).
				Padding(0, 1)

	// Post Detail Styles - Enhanced
	styleDetailBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colPurple).
			Padding(1, 2).
			MarginTop(1)

	styleMetaLabel = lipgloss.NewStyle().
			Foreground(colOrange).
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
			Border(lipgloss.RoundedBorder()).
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
				Border(lipgloss.ThickBorder()).
				BorderForeground(colCyan).
				Padding(1, 2).
				Width(80)
)
