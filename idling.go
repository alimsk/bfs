package main

import (
	"time"

	"github.com/alimsk/shopee"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type IdlingModel struct {
	spinner  spinner.Model
	c        shopee.Client
	item     shopee.CheckoutableItem
	payment  shopee.PaymentChannelData
	addr     shopee.AddressInfo
	logistic shopee.LogisticChannelInfo
	status   string
	msgch    <-chan tea.Msg
	time     string
	err      error
	win      tea.WindowSizeMsg
	spent    time.Duration
	loading  bool
}

func NewIdlingModel(
	c shopee.Client,
	item shopee.CheckoutableItem,
	payment shopee.PaymentChannelData,
	addr shopee.AddressInfo,
	logistic shopee.LogisticChannelInfo,
) IdlingModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return IdlingModel{
		spinner:  sp,
		c:        c,
		item:     item,
		payment:  payment,
		addr:     addr,
		logistic: logistic,
		status:   "Menunggu flash sale",
		time:     time.Now().Format(timeFormat),
		loading:  true,
	}
}

const timeFormat = "3:04:05 PM"

type timeMsg string

func updateTime() tea.Msg {
	time.Sleep(time.Second - time.Since(time.Now().Round(time.Second)))
	return timeMsg(time.Now().Format(timeFormat))
}

func (m IdlingModel) Init() tea.Cmd {
	return tea.Batch(waitForFS(m.item), m.spinner.Tick, updateTime)
}

func (m IdlingModel) View() string {
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

type checkoutMsg struct{}

func waitForFS(item shopee.CheckoutableItem) tea.Cmd {
	return func() tea.Msg {
		if !item.IsFlashSale() {
			fsaletime := time.Unix(item.UpcomingFsaleStartTime(), 0)
			time.Sleep(time.Until(fsaletime))
		}
		return checkoutMsg{}
	}
}

type statusMsg string
type checkoutResultMsg struct{ spent time.Duration }

func (m IdlingModel) checkout() <-chan tea.Msg {
	ch := make(chan tea.Msg)
	go func() {
		start := time.Now()
		ch <- statusMsg(blueStyle.Render("Validasi checkout"))
		err := m.c.ValidateCheckout(m.item)
		if err != nil {
			ch <- err
			close(ch)
			return
		}
		ch <- statusMsg(blueStyle.Render("Checkout get"))
		params, err := m.c.CheckoutGetQuick(shopee.CheckoutParams{
			Addr:        m.addr,
			Item:        m.item,
			PaymentData: m.payment,
			Logistic:    m.logistic,
		})
		if err != nil {
			ch <- err
			close(ch)
			return
		}
		ch <- statusMsg(blueStyle.Render("Place order"))
		err = m.c.PlaceOrder(params)
		if err != nil {
			ch <- err
			close(ch)
			return
		}
		spent := time.Since(start)
		ch <- checkoutResultMsg{spent}
		close(ch)
	}()
	return ch
}

func waitForMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

func (m IdlingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeMsg:
		m.time = string(msg)
		return m, updateTime
	case checkoutMsg:
		m.msgch = m.checkout()
		return m, waitForMsg(m.msgch)
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
