package main

import (
	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/list"
	"github.com/alimsk/shopee"
	tea "github.com/charmbracelet/bubbletea"
)

type PaymentModel struct {
	c      shopee.Client
	item   shopee.CheckoutableItem
	list   list.Model
	opts   list.Model
	win    tea.WindowSizeMsg
	hasopt bool
}

func NewPaymentModel(c shopee.Client, item shopee.CheckoutableItem) PaymentModel {
	a := make(SingleLineAdapter, len(shopee.PaymentChannelList))
	for i, p := range shopee.PaymentChannelList {
		a[i] = [2]string{"> ", p.Name}
	}
	l := list.New(a)
	l.Focus()
	l.VisibleItemCount = 4
	return PaymentModel{
		c:    c,
		item: item,
		list: l,
	}
}

func (PaymentModel) Init() tea.Cmd { return nil }

func (m PaymentModel) View() string {
	var content string
	if m.hasopt {
		content = m.opts.View()
	} else {
		content = m.list.View()
	}

	return bold("Pilih metode pembayaran") + "\n\n" +
		warnStyle.Copy().
			Width(m.win.Width-1).
			Render("Note: beberapa metode pembayaran mungkin tidak tersedia, namun tetap ditampilkan") + "\n\n" +
		content
}

func (m PaymentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "w":
			m.list.SetItemFocus(m.list.ItemFocus() - 1)
		case "s":
			m.list.SetItemFocus(m.list.ItemFocus() + 1)
		case "enter":
			p := shopee.PaymentChannelList[m.list.ItemFocus()]

			if m.hasopt {
				pdata := p.ApplyOpt(p.Options[m.opts.ItemFocus()])
				return m, navigator.PushReplacement(NewLogisticModel(m.c, m.item, pdata))
			}

			if opts := shopee.PaymentChannelList[m.list.ItemFocus()].Options; len(opts) != 0 {
				a := make(SingleLineAdapter, len(opts))
				for i, opt := range opts {
					a[i] = [2]string{"> ", opt.Name}
				}
				m.opts = list.New(a)
				m.opts.Focus()
				m.list.Blur()
				m.hasopt = true
				return m, nil
			}

			return m, navigator.PushReplacement(NewLogisticModel(m.c, m.item, p.Apply()))
		case "esc":
			if m.hasopt {
				m.opts.Blur()
				m.list.Focus()
				m.hasopt = false
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.win = msg
	}

	var cmd1, cmd2 tea.Cmd
	m.list, cmd1 = m.list.Update(msg)
	m.opts, cmd2 = m.opts.Update(msg)
	return m, tea.Batch(cmd1, cmd2)
}
