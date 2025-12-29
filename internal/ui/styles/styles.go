package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - Nord-inspired with some custom tweaks
var (
	// Primary colors
	Primary       = lipgloss.Color("#88C0D0")
	PrimaryDark   = lipgloss.Color("#5E81AC")
	Secondary     = lipgloss.Color("#81A1C1")
	Accent        = lipgloss.Color("#8FBCBB")
	AccentBright  = lipgloss.Color("#A3BE8C")
	Warning       = lipgloss.Color("#EBCB8B")
	Error         = lipgloss.Color("#BF616A")
	Success       = lipgloss.Color("#A3BE8C")

	// Background colors
	BgDark       = lipgloss.Color("#2E3440")
	BgMedium     = lipgloss.Color("#3B4252")
	BgLight      = lipgloss.Color("#434C5E")
	BgHighlight  = lipgloss.Color("#4C566A")

	// Text colors
	TextPrimary   = lipgloss.Color("#ECEFF4")
	TextSecondary = lipgloss.Color("#D8DEE9")
	TextMuted     = lipgloss.Color("#9099A1")
	TextDim       = lipgloss.Color("#6C7A89")

	// Special colors
	Border       = lipgloss.Color("#4C566A")
	BorderActive = lipgloss.Color("#88C0D0")
	Cursor       = lipgloss.Color("#D08770")
)

// Text styles
var (
	TextMutedStyle = lipgloss.NewStyle().Foreground(TextMuted)
	TextDimStyle   = lipgloss.NewStyle().Foreground(TextDim)
)

// Base styles
var (
	// App container
	AppStyle = lipgloss.NewStyle().
			Background(BgDark)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Background(BgMedium).
			Padding(0, 2).
			MarginBottom(1)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Background(BgMedium)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			MarginBottom(1)

	// Box/Panel styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderActive).
			Padding(1, 2)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(BgHighlight).
				Bold(true).
				PaddingLeft(2)

	ActiveItemStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			PaddingLeft(1)

	// Menu styles
	MenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2).
			Width(30)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			PaddingLeft(2)

	MenuSelectedStyle = lipgloss.NewStyle().
				Foreground(BgDark).
				Background(Primary).
				Bold(true).
				PaddingLeft(2).
				PaddingRight(2)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgLight).
			Bold(true).
			Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			Background(BgMedium).
			Padding(0, 1).
			MarginBottom(1)

	// Form styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Width(20)

	InputStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgLight).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(BgHighlight).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(Primary).
				Padding(0, 1)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgLight).
			Padding(0, 3).
			MarginRight(1)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(BgDark).
				Background(Primary).
				Bold(true).
				Padding(0, 3).
				MarginRight(1)

	ButtonDangerStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(Error).
				Bold(true).
				Padding(0, 3).
				MarginRight(1)

	ButtonSuccessStyle = lipgloss.NewStyle().
				Foreground(BgDark).
				Background(Success).
				Bold(true).
				Padding(0, 3).
				MarginRight(1)

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgMedium).
			Padding(0, 2)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgLight).
			Padding(0, 1).
			MarginRight(1)

	StatusDescStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			MarginRight(2)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	// Code/Editor styles
	CodeStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgMedium)

	LineNumberStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			Width(4).
			Align(lipgloss.Right).
			MarginRight(1)

	// Category/Group styles
	CategoryStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	CategoryIconStyle = lipgloss.NewStyle().
				Foreground(Accent).
				MarginRight(1)

	// Property styles
	PropertyNameStyle = lipgloss.NewStyle().
				Foreground(Primary)

	PropertyValueStyle = lipgloss.NewStyle().
				Foreground(TextSecondary)

	PropertyTypeStyle = lipgloss.NewStyle().
				Foreground(TextMuted).
				Italic(true)

	// Tooltip/Description styles
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(TextMuted).
				Italic(true).
				PaddingLeft(2)

	// Tag styles
	TagStyle = lipgloss.NewStyle().
			Foreground(BgDark).
			Background(Secondary).
			Padding(0, 1).
			MarginRight(1)

	TagActiveStyle = lipgloss.NewStyle().
			Foreground(BgDark).
			Background(AccentBright).
			Padding(0, 1).
			MarginRight(1)

	// Modal styles
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(Primary).
			Background(BgMedium).
			Padding(2, 4)

	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			MarginBottom(1)

	// Notification styles
	NotifyInfoStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(PrimaryDark).
			Padding(0, 2)

	NotifySuccessStyle = lipgloss.NewStyle().
				Foreground(BgDark).
				Background(Success).
				Padding(0, 2)

	NotifyWarningStyle = lipgloss.NewStyle().
				Foreground(BgDark).
				Background(Warning).
				Padding(0, 2)

	NotifyErrorStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(Error).
				Padding(0, 2)

	// Preview styles
	PreviewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	PreviewHeaderStyle = lipgloss.NewStyle().
				Foreground(Accent).
				Bold(true).
				MarginBottom(1)

	// Empty state style
	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true).
			Align(lipgloss.Center)
)

// Icons and symbols
const (
	IconArrowRight  = "→"
	IconArrowLeft   = "←"
	IconArrowUp     = "↑"
	IconArrowDown   = "↓"
	IconCheck       = "✓"
	IconCross       = "✗"
	IconDot         = "●"
	IconCircle      = "○"
	IconSquare      = "■"
	IconTriangle    = "▶"
	IconStar        = "★"
	IconHeart       = "♥"
	IconWarning     = "⚠"
	IconInfo        = "ℹ"
	IconEdit        = "✎"
	IconSave        = "💾"
	IconFolder      = "📁"
	IconFile        = "📄"
	IconGear        = "⚙"
	IconPalette     = "🎨"
	IconPlugin      = "🔌"
	IconRefresh     = "↻"
	IconPlus        = "+"
	IconMinus       = "-"
	IconSelected    = "▸"
	IconNotSelected = " "

	// Box drawing
	BoxTopLeft     = "╭"
	BoxTopRight    = "╮"
	BoxBottomLeft  = "╰"
	BoxBottomRight = "╯"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
)

// Helper functions
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Truncate truncates a string to a maximum length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Pad pads a string to a minimum length
func Pad(s string, minLen int) string {
	if len(s) >= minLen {
		return s
	}
	return s + string(make([]byte, minLen-len(s)))
}

// Center centers a string within a given width
func Center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	return string(make([]byte, padding)) + s + string(make([]byte, width-len(s)-padding))
}
