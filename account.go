package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"

	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/list"
	"github.com/alimsk/shopee"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	jsoniter "github.com/json-iterator/go"
)

type CookieInputModel struct {
	spinner spinner.Model
	input   textinput.Model
	state   *State
	win     tea.WindowSizeMsg
	err     error
	loading bool
}

func NewCookieInputModel(s *State) CookieInputModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	i := textinput.New()
	i.Focus()
	i.Placeholder = "Masukkan cookie"
	i.TextStyle = focusedStyle
	i.CursorStyle = focusedStyle
	i.PromptStyle = focusedStyle
	return CookieInputModel{
		spinner: sp,
		input:   i,
		state:   s,
	}
}

func (m CookieInputModel) Init() tea.Cmd { return tea.Batch(textinput.Blink, m.spinner.Tick) }

func (m CookieInputModel) View() string {
	var b strings.Builder
	b.WriteString(bold("Login dengan cookie") + "\n\n")
	if m.loading {
		b.WriteString(m.spinner.View() + "Loading...")
	} else {
		b.WriteString(m.input.View())
	}
	if m.err != nil {
		b.WriteString("\n\n" + errorStyle.Copy().Width(m.win.Width-1).Render("error: "+m.err.Error()) + "\n")
	}
	return b.String()
}

func (m CookieInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.loading = true
			m.input.Blur()
			return m, login([]byte(m.input.Value()))
		case "esc":
			return m, navigator.Pop()
		}
	case loginResultMsg:
		m.loading = false
		m.input.Focus()
		if msg.err != nil {
			m.input.SetValue("")
			m.err = msg.err
			return m, nil
		}
		m.state.Cookies = append([]*CookieJarMarshaler{{msg.c.Client.GetClient().Jar}}, m.state.Cookies...)
		return m, navigator.PushAndRemoveUntil(
			NewAccountModel(m.state),
			func(int, tea.Model) bool { return false },
		)
	case error:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.win = msg
	}

	var cmd1, cmd2 tea.Cmd
	m.input, cmd1 = m.input.Update(msg)
	m.spinner, cmd2 = m.spinner.Update(msg)
	return m, tea.Batch(cmd1, cmd2)
}

type AccountModel struct {
	list         list.Model
	spinner      spinner.Model
	state        *State
	err          error
	shortcuthelp string
	cs           []shopee.Client
	win          tea.WindowSizeMsg
	initialized  bool
}

func NewAccountModel(s *State) AccountModel {
	l := list.New(SingleLineAdapter{{"+ ", "Login"}})
	l.Focus()
	l.VisibleItemCount = 4
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return AccountModel{
		state:   s,
		list:    l,
		spinner: sp,
		shortcuthelp: fmt.Sprint(
			keyhelp("↑", "move up"), keysep, keyhelp("↓", "move down"), "\n",
			keyhelp("Enter", "choose account"),
		),
	}
}

type accountChooserInitMsg struct {
	cs      []shopee.Client
	usernms []string
	cookies []*CookieJarMarshaler
}

func (m AccountModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			cs := make([]shopee.Client, 0, len(m.state.Cookies))
			usernms := make([]string, 0, len(m.state.Cookies))
			newcookies := m.state.Cookies[:0]
			for _, cookie := range m.state.Cookies {
				c, err := shopee.New(cookie.CookieJar)
				if err != nil {
					return err
				}
				acc, err := c.FetchAccountInfo()
				if err != nil {
					continue
				}
				cs = append(cs, c)
				usernms = append(usernms, acc.Username())
				newcookies = append(newcookies, cookie)
			}

			return accountChooserInitMsg{cs, usernms, newcookies}
		},
	)
}

func (m AccountModel) View() string {
	var b strings.Builder
	b.WriteString(bold("Pilih Akun") + "\n\n")
	if m.initialized {
		b.WriteString(m.list.View())
	} else {
		b.WriteString(m.spinner.View() + "Loading...")
	}

	if m.err != nil {
		b.WriteString("\n\n" + errorStyle.Copy().Width(m.win.Width-1).Render("error: "+m.err.Error()))
	}

	b.WriteString("\n\n" + m.shortcuthelp)

	return b.String()
}

type loginResultMsg struct {
	c   shopee.Client
	acc shopee.AccountInfo
	err error
}

func login(cookie []byte) tea.Cmd {
	return func() tea.Msg {
		if !jsoniter.Valid(cookie) {
			return loginResultMsg{err: errors.New("not a valid json input")}
		}

		json := jsoniter.Get(cookie)
		cookies := make([]*http.Cookie, json.Size())
		for i := 0; i < json.Size(); i++ {
			item := json.Get(i)
			value, err := strconv.Unquote(item.Get("value").ToString())
			if err != nil {
				value = item.Get("value").ToString()
			}
			// do not set expires
			cookies[i] = &http.Cookie{
				Name:   item.Get("name").ToString(),
				Value:  value,
				Domain: item.Get("domain").ToString(),
				// Expires:  time.Unix(item.Get("expirationDate").ToInt64(), 0),
				HttpOnly: item.Get("httpOnly").ToBool(),
				Path:     item.Get("path").ToString(),
				Secure:   item.Get("secure").ToBool(),
			}
		}

		jar, _ := cookiejar.New(nil)
		jar.SetCookies(shopee.ShopeeUrl, cookies)
		c, err := shopee.New(jar)
		if err != nil {
			return loginResultMsg{err: err}
		}
		acc, err := c.FetchAccountInfo()
		return loginResultMsg{c, acc, err}
	}
}

func (m AccountModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.initialized {
			// do not accept user input on initialization process
			break
		}
		switch msg.String() {
		case "w":
			m.list.SetItemFocus(m.list.ItemFocus() - 1)
		case "s":
			m.list.SetItemFocus(m.list.ItemFocus() + 1)
		case "enter":
			if m.list.ItemFocus() == m.list.Adapter.Len()-1 {
				m.err = nil
				return m, navigator.Push(NewCookieInputModel(m.state))
			}
			usernm := m.list.Adapter.(SingleLineAdapter)[m.list.ItemFocus()][1]
			return m, navigator.PushReplacement(NewURLModel(m.cs[m.list.ItemFocus()], usernm))
		}
	case accountChooserInitMsg:
		m.cs = msg.cs
		m.state.Cookies = msg.cookies
		a := make(SingleLineAdapter, len(msg.usernms)+1)
		for i, usernm := range msg.usernms {
			a[i] = [2]string{"> ", usernm}
		}
		a[len(a)-1] = m.list.Adapter.(SingleLineAdapter)[0]
		m.list.Adapter = a
		m.initialized = true
	case loginResultMsg:
		m.list.Focus()
		if msg.err != nil {
			m.err = msg.err
			break
		}
		m.cs = append([]shopee.Client{msg.c}, m.cs...)
		m.state.Cookies = append([]*CookieJarMarshaler{{msg.c.Client.GetClient().Jar}}, m.state.Cookies...)
		m.list.Adapter = append(SingleLineAdapter{[2]string{"> ", msg.acc.Username()}}, m.list.Adapter.(SingleLineAdapter)...)
		m.err = m.state.saveAsFile(*stateFilename)
	case error:
		m.err = msg
	case tea.WindowSizeMsg:
		m.win = msg
	}

	var cmd1, cmd2 tea.Cmd
	m.spinner, cmd1 = m.spinner.Update(msg)
	m.list, cmd2 = m.list.Update(msg)
	return m, tea.Batch(cmd1, cmd2)
}
