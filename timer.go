package main

import (
	"time"

	"github.com/alimsk/shopee"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TimerModel struct {
	spinner  spinner.Model
	c        shopee.Client
	item     shopee.CheckoutableItem
	payment  shopee.PaymentChannelData
	addr     shopee.AddressInfo
	logistic shopee.LogisticChannelInfo
	status   string
	msgch    chan tea.Msg
	time     string
	err      error
	win      tea.WindowSizeMsg
	spent    time.Duration
	loading  bool
}

func NewTimerModel(
	c shopee.Client,
	item shopee.CheckoutableItem,
	payment shopee.PaymentChannelData,
	addr shopee.AddressInfo,
	logistic shopee.LogisticChannelInfo,
) TimerModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return TimerModel{
		spinner:  sp,
		c:        c,
		item:     item,
		payment:  payment,
		addr:     addr,
		logistic: logistic,
		status:   "Menunggu flash sale",
		time:     time.Now().Format(timeFormat),
		loading:  true,
		msgch:    make(chan tea.Msg),
	}
}

const timeFormat = "3:04:05 PM"

type timeMsg string

func updateTime() tea.Msg {
	time.Sleep(time.Second - time.Since(time.Now().Round(time.Second)))
	return timeMsg(time.Now().Format(timeFormat))
}

func (m TimerModel) Init() tea.Cmd {
	go m.checkout()
	return tea.Batch(waitForMsg(m.msgch), m.spinner.Tick, updateTime)
}

func (m TimerModel) View() string {
	if m.err != nil {
		return m.status + "\n" +
			errorStyle.Copy().
				Width(m.win.Width-1).
				Render(m.err.Error()) + "\n" // trailing line prevent from erasing last line
	}

	var spinnerview string
	if m.loading {
		spinnerview = m.spinner.View() + " "
	}
	var spent string
	if m.spent != 0 {
		spent = ternary(m.spent.Seconds() < 2, successStyle, warnStyle).Render(m.spent.String())
	}
	return lipgloss.NewStyle().
		Width(25).
		Align(lipgloss.Center).
		PaddingLeft(2).
		Render(
			blueStyle.Render(m.time)+"\n"+
				spinnerview+m.status+"\n"+
				spent,
		) + "\n"
}

type statusMsg string
type checkoutResultMsg struct{ spent time.Duration }

func (m TimerModel) checkout() {
	if !m.item.IsFlashSale() {
		fsaletime := time.Unix(m.item.UpcomingFsaleStartTime(), 0)
		time.Sleep(time.Until(fsaletime))
	}

	start := time.Now()
	m.msgch <- statusMsg(blueStyle.Render("Refreshing item"))
	updateditem, err := m.c.FetchItem(m.item.ShopID(), m.item.ItemID())
	if err != nil {
		m.msgch <- err
		close(m.msgch)
		return
	}
	m.msgch <- statusMsg(blueStyle.Render("Validasi checkout"))
	citem := shopee.ChooseModel(updateditem, m.item.ChosenModel().ModelID())
	err = m.c.ValidateCheckout(citem)
	if err != nil {
		if coErr, ok := err.(shopee.CheckoutValidationError); !(ok && coErr.Code() == 7) {
			m.msgch <- err
			close(m.msgch)
			return
		}
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
	case timeMsg:
		m.time = string(msg)
		return m, updateTime
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

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}
