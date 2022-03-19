package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/alimsk/bfs/navigator"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	stateFilename = flag.String("state", "state.json", "state file name")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	state, err := loadStateFile(*stateFilename)
	if errors.Is(err, os.ErrNotExist) {
		state = &State{}
	} else if err != nil {
		log.Fatal(err)
	}
	defer state.saveAsFile(*stateFilename)

	m := navigator.New(NewAccountModel(state))
	p := tea.NewProgram(m)
	if err = p.Start(); err != nil {
		log.Fatal(err)
	}

	if runtime.GOOS == "windows" {
		fmt.Scanln()
	}
}
