package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/shopee"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ItemModel struct {
	tvars []shopee.TierVar
	// currently focused option
	tvarfocus []int
	citem     shopee.CheckoutableItem
	c         shopee.Client
	item      shopee.Item
	err       error
	focus     int
	win       tea.WindowSizeMsg
}

func NewItemModel(c shopee.Client, item shopee.Item) ItemModel {
	tvars := item.TierVariations()
	tvarfocus := make([]int, len(tvars))
	return ItemModel{
		item:      item,
		tvars:     tvars,
		tvarfocus: tvarfocus,
		citem:     shopee.ChooseModelByTierVar(item, tvarfocus),
		c:         c,
		focus:     ternary(hasNoVariant(tvars), len(tvars), 0),
	}
}

func (m ItemModel) Init() tea.Cmd { return nil }

func (m ItemModel) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().
		Width(m.win.Width-2).
		Border(lipgloss.NormalBorder(), true).
		Padding(0, 1).
		Render(
			blueStyle.Render(m.item.Name()) + "\n" +
				bold("Harga: ") + blueStyle.Render(formatPrice(m.item.Price())) + "\n" +
				bold("Stok:  ") + blueStyle.Render(strconv.Itoa(m.item.Stock())),
		),
	)
	b.WriteByte('\n')
	cm := m.citem.ChosenModel()
	b.WriteString(lipgloss.NewStyle().
		Width(m.win.Width-2).
		Border(lipgloss.NormalBorder(), true).
		Padding(0, 1).
		Render(
			bold("Model") + "\n" +
				blueStyle.Render(ternary(cm.Name() == "", "Tidak ada varian", cm.Name())) + "\n" +
				bold("Harga: ") + blueStyle.Render(formatPrice(cm.Price())) + "\n" +
				bold("Stok:  ") + ternary(cm.Stock() != 0, blueStyle, errorStyle).Render(strconv.Itoa(cm.Stock())),
		),
	)
	b.WriteByte('\n')
	alignright := lipgloss.NewStyle().
		Width(m.win.Width / 2).
		Align(lipgloss.Right).
		PaddingRight(2)
	alignleft := lipgloss.NewStyle().
		Width(m.win.Width / 2).
		Align(lipgloss.Left).
		PaddingLeft(2)
	for i, tvar := range m.tvars {
		name := tvar.Name()
		if m.focus == i {
			name = blueStyle.Render(name)
		}
		b.WriteString(alignleft.Render(name))
		opt := tvar.Options()[m.tvarfocus[i]]
		ri := ternary(m.tvarfocus[i] == len(m.tvars[i].Options())-1, "  ", " >")
		li := ternary(m.tvarfocus[i] == 0, "  ", "< ")
		render := ternary(m.focus == i, focusedStyle, blurredStyle).Render
		b.WriteString(alignright.Render(render(li + opt + ri)))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')

	confirm := ternary(m.focus == len(m.tvars), focusedStyle, blurredStyle).Render
	b.WriteString(alignright.
		Width(m.win.Width).
		Render(confirm("[ Next ]")),
	)

	if m.err != nil {
		b.WriteString("\n" + errorStyle.Copy().Width(m.win.Width-1).Render("error: "+m.err.Error()) + "\n")
	}

	return b.String()
}

func (m ItemModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focus == len(m.tvars) {
				if !m.citem.HasUpcomingFsale() && m.citem.ChosenModel().Stock() == 0 {
					m.err = errors.New("stok model kosong")
					return m, nil
				}
				return m, navigator.PushReplacement(NewPaymentModel(m.c, m.citem))
			} else {
				m.focus = min(len(m.tvars), m.focus+1)
			}
		case "up", "w", "shift+tab":
			if hasNoVariant(m.tvars) {
				return m, nil
			}
			m.focus = max(0, m.focus-1)
		case "down", "s", "tab":
			if hasNoVariant(m.tvars) {
				return m, nil
			}
			m.focus = min(len(m.tvars), m.focus+1)
		case "left", "d":
			if m.focus < len(m.tvars) {
				m.tvarfocus[m.focus] = max(0, m.tvarfocus[m.focus]-1)
				m.citem = shopee.ChooseModelByTierVar(m.item, m.tvarfocus)
			}
		case "right", "a":
			if m.focus < len(m.tvars) {
				m.tvarfocus[m.focus] = min(len(m.tvars[m.focus].Options())-1, m.tvarfocus[m.focus]+1)
				m.citem = shopee.ChooseModelByTierVar(m.item, m.tvarfocus)
			}
		}
	case tea.WindowSizeMsg:
		m.win = msg
	}
	return m, nil
}

func hasNoVariant(tvars []shopee.TierVar) bool {
	return len(tvars) == 1 && len(tvars[0].Options()) == 1
}
