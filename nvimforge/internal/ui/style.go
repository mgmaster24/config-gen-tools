// Package ui handles nvimforge's own terminal presentation: the CLI
// startup banner and the interactive setup prompts. It makes no decisions
// of its own — callers decide what to ask and what to do with the answers.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	bannerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
)
