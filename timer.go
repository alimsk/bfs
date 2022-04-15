package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alimsk/shopee"
	tea "github.com/charmbracelet/bubbletea"
)

type TaskStatus int

const (
	statusPending TaskStatus = iota
	statusRunning
	statusDone
)

type Task struct {
	title string

	// success = status == statusDone && err != nil
	err error

	status TaskStatus
}

type TimerModel struct {
	c        shopee.Client
	item     shopee.CheckoutableItem
	payment  shopee.PaymentChannelData
	addr     shopee.AddressInfo
	logistic shopee.LogisticChannelInfo

	fsale         time.Time
	countdownCmd  tea.Cmd
	countdownView string

	msgch       chan tea.Msg
	err         error
	tasks       []Task
	currentTask int
	spent       time.Duration

	win tea.WindowSizeMsg
}

func NewTimerModel(
	c shopee.Client,
	item shopee.CheckoutableItem,
	payment shopee.PaymentChannelData,
	addr shopee.AddressInfo,
	logistic shopee.LogisticChannelInfo,
) *TimerModel {
	fsale := time.Unix(item.UpcomingFsaleStartTime(), 0)
	return &TimerModel{
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
			"00:00:00",
		),
		msgch: make(chan tea.Msg, 1),
		tasks: []Task{
			{title: "Refreshing Item"},
			{title: "Validasi Checkout"},
			{title: "Checkout get"},
			{title: "Place order"},
		},
	}
}

type countdownMsg int64

func countdownFormat(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func countdown(fsale time.Time) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second - time.Since(time.Now().Round(time.Second)))
		d := fsale.Sub(time.Now().Local())
		return countdownMsg(d)
	}
}

func (m *TimerModel) Init() tea.Cmd {
	go m.checkout()
	return tea.Batch(waitForMsg(m.msgch), m.countdownCmd)
}

func (m *TimerModel) View() string {
	var b strings.Builder

	b.WriteString("Mulai pada " + blueStyle.Render(m.countdownView) + "\n")
	for _, task := range m.tasks {
		var cursor string
		var style func(string) string
		switch task.status {
		case statusPending:
			cursor = "[ ] "
			style = blurredStyle.Render
		case statusRunning:
			cursor = "[*] "
			style = blueStyle.Render
		case statusDone:
			if task.err != nil {
				cursor = "[𐄂] "
				style = errorStyle.Render
			} else {
				cursor = "[✓] "
				style = successStyle.Render
			}
		}
		b.WriteString(style(cursor+task.title) + "\n")
	}

	if m.err != nil {
		b.WriteString("\n" +
			errorStyle.Copy().
				Width(m.win.Width-1).
				Render(m.err.Error()) + "\n",
		) // trailing line prevent from erasing last line
	} else if m.spent != 0 {
		// show this message only if m.err == nil
		b.WriteString("\nSukses dalam ")
		b.WriteString(ternary(m.spent.Seconds() < 2, successStyle, warnStyle).Render(m.spent.String()))
	}

	return b.String() + "\n"
}

type taskUpdateMsg struct{ status TaskStatus }
type checkoutResultMsg struct{ spent time.Duration }

func (m *TimerModel) checkout() {
	start := time.Now()

	updateditem := m.item.Item
	if !m.item.Item.IsFlashSale() {
		time.Sleep(time.Until(m.fsale))
		m.msgch <- taskUpdateMsg{statusRunning}
		start = time.Now()
		var err error
		updateditem, err = m.c.FetchItem(m.item.ShopID(), m.item.ItemID())
		if err != nil {
			m.msgch <- err
			close(m.msgch)
			return
		}
	}
	m.msgch <- taskUpdateMsg{statusDone}

	var wg sync.WaitGroup
	wg.Add(2)

	citem := shopee.ChooseModel(updateditem, m.item.ChosenModel().ModelID())
	go func() {
		m.msgch <- taskUpdateMsg{statusRunning}
		err := m.c.ValidateCheckout(citem)
		if err != nil {
			m.msgch <- err
			close(m.msgch)
			return
		}
		m.msgch <- taskUpdateMsg{statusDone}

		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)

	go func() {
		m.msgch <- taskUpdateMsg{statusRunning}
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
		m.msgch <- taskUpdateMsg{statusDone}

		m.msgch <- taskUpdateMsg{statusRunning}
		err = m.c.PlaceOrder(params)
		if err != nil {
			m.msgch <- err
			close(m.msgch)
			return
		}
		m.msgch <- taskUpdateMsg{statusDone}

		wg.Done()
	}()

	wg.Wait()

	spent := time.Since(start)
	m.msgch <- checkoutResultMsg{spent}
	close(m.msgch)
}

func waitForMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

func (m *TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case countdownMsg:
		if msg <= 0 {
			m.countdownView = "00:00:00"
			return m, nil
		}
		m.countdownView = countdownFormat(time.Duration(msg))
		return m, m.countdownCmd
	case checkoutResultMsg:
		m.spent = msg.spent
		return m, tea.Quit
	case taskUpdateMsg:
		m.tasks[m.currentTask].status = msg.status
		switch msg.status {
		case statusDone:
			m.currentTask++
		case statusPending, statusRunning:
			// do nothing, update views
		}
		return m, waitForMsg(m.msgch)
	case error:
		m.err = msg
		m.tasks[m.currentTask].err = msg
		m.tasks[m.currentTask].status = statusDone
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.win = msg
	}

	return m, nil
}
