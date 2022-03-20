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
	platform      = flag.String("platform", "web",
		"choose platform, currently supported platforms are:\n"+
			"  web\n"+
			"  android\n"+
			"in most cases, web will work more than others",
	)
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	switch *platform {
	case "web", "android":
	default:
		log.Fatal("unknown platform:", *platform)
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
