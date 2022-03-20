package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/alimsk/bfs/navigator"
	"github.com/alimsk/shopee"
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

func platformOption() shopee.Option {
	switch *platform {
	case "web":
		return shopee.WithWeb
	case "android":
		return shopee.WithAndroidApp
	}
	return nil
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if platformOption() == nil {
		log.Fatal("unknown platform: ", *platform)
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
