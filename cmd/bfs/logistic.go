package main

import (
	"errors"
	"strings"

	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/list"
	"github.com/alimsk/shopee"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type LogisticModel struct {
	c             shopee.Client
	item          shopee.CheckoutableItem
	payment       shopee.PaymentChannel
	paymentOption string
	addr          shopee.AddressInfo

	spinner   spinner.Model
	list      list.Model
	win       tea.WindowSizeMsg
	err       error
	logistics []shopee.LogisticChannelInfo
}

func NewLogisticModel(c shopee.Client, item shopee.CheckoutableItem, payment shopee.PaymentChannel, paymentOption string) LogisticModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return LogisticModel{
		spinner: sp,
		c:       c,
		item:    item,
		payment: payment,
	}
}

type logisticInitMsg struct {
	addr      shopee.AddressInfo
	logistics []shopee.LogisticChannelInfo
}

type fatalError struct{ error }

func (m LogisticModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			addrs, err := m.c.FetchAddresses()
			if err != nil {
				return err
			}
			i, deliveryAddr := addrs.DeliveryAddress()
			if i == -1 {
				return fatalError{errors.New("alamat utama tidak ditemukan, silahkan setting terlebih dahulu")}
			}

			logistics, err := m.c.FetchShippingInfo(deliveryAddr, m.item.Item)
			if err != nil {
				return err
			}
			return logisticInitMsg{deliveryAddr, logistics}
		},
	)
}

func (m LogisticModel) View() string {
	var b strings.Builder
	b.WriteString(bold("Pilih channel logistik") + "\n\n")
	if m.logistics == nil {
		b.WriteString(m.spinner.View() + "Loading...")
	} else {
		b.WriteString(m.list.View())
	}
	b.WriteString("\n\n")
	if m.err != nil {
		b.WriteString(errorStyle.Copy().Width(m.win.Width-1).Render("error: "+m.err.Error()) + "\n")
	}
	return b.String()
}

func (m LogisticModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// interfaces must be placed below concrete types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.logistics == nil {
			// is fetching
			return m, nil
		}
		switch msg.String() {
		case "w":
			m.list.SetItemFocus(m.list.ItemFocus() - 1)
		case "s":
			m.list.SetItemFocus(m.list.ItemFocus() + 1)
		case "enter":
			lc := m.logistics[m.list.ItemFocus()]
			if m.list.Adapter.(*list.SimpleAdapter).ItemAt(m.list.ItemFocus()).Disabled {
				return m, nil
			}
			return m, navigator.PushAndRemoveUntil(
				NewTimerModel(m.c, m.item, m.payment, m.paymentOption, m.addr, lc),
				func(int, tea.Model) bool { return false },
			)
		}
	case logisticInitMsg:
		m.logistics = msg.logistics
		m.addr = msg.addr
		items := make(list.SimpleItemList, len(msg.logistics))
		for i, logistic := range msg.logistics {
			var desc string
			if logistic.HasWarning() {
				desc = logistic.Warning()
			} else {
				if logistic.MinPrice() == logistic.MaxPrice() {
					desc = formatPrice(logistic.MaxPrice())
				} else {
					desc = formatPrice(logistic.MinPrice()) + " - " + formatPrice(logistic.MaxPrice())
				}
				// if logistic.PriceBeforeDiscount() != 0 {
				// 	desc += " " + blurredStyle.Copy().Strikethrough(true).Render(formatPrice(logistic.PriceBeforeDiscount()))
				// }
			}
			items[i] = list.SimpleItem{
				Title:    logistic.Name(),
				Desc:     desc,
				Disabled: logistic.HasWarning(),
			}
		}
		a := list.NewSimpleAdapter(items)
		m.list = list.New(a)
		m.list.Focus()
		if len(msg.logistics) == 1 {
			if msg.logistics[0].HasWarning() {
				m.err = errors.New("tidak ada channel logistik tersedia")
				return m, tea.Quit
			}
			return m, navigator.PushAndRemoveUntil(
				NewTimerModel(m.c, m.item, m.payment, m.paymentOption, msg.addr, msg.logistics[0]),
				func(int, tea.Model) bool { return false },
			)
		}
	case fatalError:
		m.err = msg.error
		return m, tea.Quit
	case error:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.win = msg
	}

	var cmd1, cmd2 tea.Cmd
	m.list, cmd1 = m.list.Update(msg)
	m.spinner, cmd2 = m.spinner.Update(msg)
	return m, tea.Batch(cmd1, cmd2)
}
