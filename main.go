package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alimsk/bfs/navigator"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	stateFilename = flag.String("state", "bfs_state.json", "state file name")
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
	if runtime.GOOS == "android" {
		fixTimezone()
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

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
