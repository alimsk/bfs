package main

import (
	"fmt"
	"time"

	"github.com/alimsk/shopee"
	tea "github.com/charmbracelet/bubbletea"
)

type TimerModel struct {
	c             shopee.Client
	item          shopee.CheckoutableItem
	payment       shopee.PaymentChannelData
	addr          shopee.AddressInfo
	logistic      shopee.LogisticChannelInfo
	fsale         time.Time
	countdownCmd  tea.Cmd
	status        string
	msgch         chan tea.Msg
	countdownView string
	err           error
	win           tea.WindowSizeMsg
	spent         time.Duration
	loading       bool
}

func NewTimerModel(
	c shopee.Client,
	item shopee.CheckoutableItem,
	payment shopee.PaymentChannelData,
	addr shopee.AddressInfo,
	logistic shopee.LogisticChannelInfo,
) TimerModel {
	fsale := time.Unix(item.UpcomingFsaleStartTime(), 0)
	return TimerModel{
		c:            c,
		item:         item,
		payment:      payment,
		addr:         addr,
		logistic:     logistic,
		fsale:        fsale,
		countdownCmd: ternary(item.HasUpcomingFsale(), countdown(fsale), nil),
		countdownView: ternary(
			item.HasUpcomingFsale(),
			countdownFormat(fsale.Sub(time.Now().Local())),
			"",
		),
		loading: true,
		msgch:   make(chan tea.Msg, 1),
	}
}

type countdownMsg string

func countdownFormat(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func countdown(fsale time.Time) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second - time.Since(time.Now().Round(time.Second)))
		d := fsale.Sub(time.Now().Local())
		if d <= 0 {
			return countdownMsg("")
		}
		return countdownMsg(countdownFormat(d))
	}
}

func (m TimerModel) Init() tea.Cmd {
	go m.checkout()
	return tea.Batch(waitForMsg(m.msgch), m.countdownCmd)
}

func (m TimerModel) View() string {
	if m.err != nil {
		return m.status + "\n" +
			errorStyle.Copy().
				Width(m.win.Width-1).
				Render(m.err.Error()) + "\n" // trailing line prevent from erasing last line
	}

	var content string
	if m.countdownView != "" {
		content = "Mulai pada " + blueStyle.Render(m.countdownView)
	} else {
		content = m.status
		if m.spent != 0 {
			content += "\n" + ternary(m.spent.Seconds() < 2, successStyle, warnStyle).Render(m.spent.String())
		}
	}

	return bold("Countdown") + "\n" +
		content + "\n"
}

type statusMsg string
type checkoutResultMsg struct{ spent time.Duration }

func (m TimerModel) checkout() {
	start := time.Now()

	updateditem := m.item.Item
	if !m.item.Item.IsFlashSale() {
		time.Sleep(time.Until(m.fsale))
		m.msgch <- statusMsg(blueStyle.Render("Refreshing item"))
		start = time.Now()
		var err error
		updateditem, err = m.c.FetchItem(m.item.ShopID(), m.item.ItemID())
		if err != nil {
			m.msgch <- err
			close(m.msgch)
			return
		}
	}

	citem := shopee.ChooseModel(updateditem, m.item.ChosenModel().ModelID())
	m.msgch <- statusMsg(blueStyle.Render("Validasi checkout"))
	err := m.c.ValidateCheckout(citem)
	if err != nil {
		m.msgch <- err
		close(m.msgch)
		return
	}
	m.msgch <- statusMsg(blueStyle.Render("Checkout get"))
	params, err := m.c.CheckoutGetQuick(shopee.CheckoutParams{
		Addr:        m.addr,
		Item:        citem,
		PaymentData: m.payment,
		Logistic:    m.logistic,
	})
	if err != nil {
		m.msgch <- err
		close(m.msgch)
		return
	}
	m.msgch <- statusMsg(blueStyle.Render("Place order"))
	err = m.c.PlaceOrder(params)
	if err != nil {
		m.msgch <- err
		close(m.msgch)
		return
	}
	spent := time.Since(start)
	m.msgch <- checkoutResultMsg{spent}
	close(m.msgch)
}

func waitForMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case countdownMsg:
		m.countdownView = string(msg)
		if msg == "" {
			return m, nil
		}
		return m, m.countdownCmd
	case checkoutResultMsg:
		m.status = successStyle.Render("Sukses")
		m.spent = msg.spent
		m.loading = false
		return m, tea.Quit
	case statusMsg:
		m.status = string(msg)
		return m, waitForMsg(m.msgch)
	case error:
		m.err = msg
		m.loading = false
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.win = msg
	}

	return m, nil
}
