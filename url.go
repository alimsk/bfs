package main

import (
	"errors"

	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/shopee"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type URLModel struct {
	input    textinput.Model
	spinner  spinner.Model
	err      error
	usernm   string
	c        shopee.Client
	win      tea.WindowSizeMsg
	fetching bool
}

func NewURLModel(c shopee.Client, usernm string) URLModel {
	i := textinput.New()
	i.Focus()
	i.Placeholder = "Masukkan URL"
	i.TextStyle = focusedStyle
	i.CursorStyle = focusedStyle
	i.PromptStyle = focusedStyle
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return URLModel{
		c:       c,
		usernm:  usernm,
		input:   i,
		spinner: sp,
	}
}

func (m URLModel) Init() tea.Cmd { return tea.Batch(textinput.Blink, m.spinner.Tick) }

func (m URLModel) View() string {
	var content string
	if m.fetching {
		content = m.spinner.View() + "Loading..."
	} else {
		content = m.input.View()
	}
	if m.err != nil {
		content += "\n\n" + errorStyle.Copy().Width(m.win.Width).Render("error: "+m.err.Error())
	}

	return bold("Masuk sebagai "+blueStyle.Render(m.usernm)) + "\n\n" +
		content + "\n\n" +
		keyhelp("ctrl+v", "paste")
}

type fetchItemMsg struct{ shopee.Item }

func (m URLModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.fetching {
				return m, nil
			}
			m.err = nil
			m.fetching = true
			return m, func() tea.Msg {
				item, err := m.c.FetchItemFromURL(m.input.Value())
				if err != nil {
					return err
				}
				if !item.IsFlashSale() && !item.HasUpcomingFsale() {
					return errors.New("tidak ada flash sale untuk item ini")
				}
				if !item.HasUpcomingFsale() && item.Stock() == 0 {
					return errors.New("stok item kosong")
				}
				return fetchItemMsg{item}
			}
		}
	case error:
		m.fetching = false
		m.input.SetValue("")
		m.err = msg
		return m, nil
	case fetchItemMsg:
		m.fetching = false
		m.input.SetValue("")
		return m, navigator.PushReplacement(NewItemModel(m.c, msg.Item))
	case tea.WindowSizeMsg:
		m.win = msg
	}

	var cmd1, cmd2 tea.Cmd
	m.input, cmd1 = m.input.Update(msg)
	m.spinner, cmd2 = m.spinner.Update(msg)
	return m, tea.Batch(cmd1, cmd2)
}
