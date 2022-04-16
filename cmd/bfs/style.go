package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	keyStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"})
	descStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"})
	blueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8fbcbb"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a3be8c"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#bf616a"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ebcb8b"))
)

var (
	border    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(0, 2).Render
	center    = lipgloss.NewStyle().Align(lipgloss.Center).Render
	bold      = lipgloss.NewStyle().Bold(true).Render
	italic    = lipgloss.NewStyle().Italic(true).Render
	underline = lipgloss.NewStyle().Underline(true).Render
)

var (
	keysep = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}).Render(" • ")
)

func keyhelp(key, desc string) string { return keyStyle.Render(key) + " " + descStyle.Render(desc) }
