package main

import "strings"

// [indicator, text]
type SingleLineAdapter [][2]string

func (s SingleLineAdapter) Len() int    { return len(s) }
func (s SingleLineAdapter) Sep() string { return "\n" }
func (s SingleLineAdapter) View(pos, focus int) string {
	if pos == focus {
		return focusedStyle.Render(s[pos][0] + s[pos][1])
	}
	return blurredStyle.Render(strings.Repeat(" ", len(s[pos][0])) + s[pos][1])
}
