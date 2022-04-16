package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/shopee"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "v1.1.0"

var (
	stateFilename = flag.String("state", "bfs_state.json", "state file name")
	delay         = flag.Duration("d", 0, "delay antar request.\ndokumentasi: https://github.com/alimsk/bfs#-d")
)

// https://github.com/golang/go/issues/20455#issuecomment-342287698
func fixTimezone() {
	out, err := exec.Command("/system/bin/getprop", "persist.sys.timezone").Output()
	if err != nil {
		return
	}
	z, err := time.LoadLocation(strings.TrimSpace(string(out)))
	if err != nil {
		return
	}
	time.Local = z
}

func init() {
	rand.Seed(time.Now().UnixNano())
	if runtime.GOOS == "android" {
		fixTimezone()
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() > 0 {
		switch flag.Arg(0) {
		case "info":
			itemInfo()
		case "version":
			fmt.Print(version)
		default:
			log.Fatal("Unknown subcommand: ", flag.Arg(0))
		}
		return
	}

	if runtime.GOOS == "windows" {
		// prevent windows auto close cmd
		defer fmt.Scanln()
	}

	state, err := loadStateFile(*stateFilename)
	if errors.Is(err, os.ErrNotExist) {
		state = &State{}
	} else if err != nil {
		log.Print(err)
		return
	}
	defer state.saveAsFile(*stateFilename)

	m := navigator.New(NewAccountModel(state))
	p := tea.NewProgram(m)
	if err = p.Start(); err != nil {
		log.Print(err)
		return
	}
}

func itemInfo() {
	urlstr := flag.Arg(1)
	printbool := func(v bool) string { return ternary(v, successStyle.Render("Ya"), errorStyle.Render("Tidak")) }

	c, err := shopee.NewFromCookieString("csrftoken=" + randstr(32))
	if err != nil {
		log.Fatal(err)
	}

	item, err := c.FetchItemFromURL(urlstr)
	if err != nil {
		log.Fatal(err)
	}

	fsalestatus := printbool(item.IsFlashSale())
	if !item.IsFlashSale() && item.HasUpcomingFsale() {
		fsalestatus = blueStyle.Render("pada jam " + time.Unix(item.UpcomingFsaleStartTime(), 0).Format("3:04:05 PM"))
	}

	m := [...]struct {
		k string
		v interface{}
	}{
		{"Flashsale", fsalestatus},
		{"Harga", formatPrice(item.Price())},
		{"Stok", item.Stock()},
		{"Kategori", strings.Join(item.CatNames(), ", ")},
		{"Shopid", item.ShopID()},
		{"Itemid", item.ItemID()},
	}

	var longestkey int
	for _, v := range m {
		if len(v.k) > longestkey {
			longestkey = len(v.k)
		}
	}

	fmt.Println(blueStyle.Render(item.Name()))
	fmt.Println()
	for _, v := range m {
		fmt.Printf("%-*s %v\n", longestkey+1, v.k+":", v.v)
	}

	for _, model := range item.Models() {
		fmt.Println(
			"\n"+blueStyle.Render(model.Name()),
			"\nID:       ", model.ModelID(),
			"\nHarga:    ", formatPrice(model.Price()),
			"\nStok:     ", model.Stock(),
			"\nFlashsale:", printbool(model.HasUpcomingFsale()),
		)
	}
}
