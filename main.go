package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var (
	timezone string
)

func main() {
	confFile := ReadConfig("config.toml")
	fmt.Println(confFile.Attributes.Path)
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/statsicon.ico"))
	timezone := "--"
	mBrowseAttributes := systray.AddMenuItem("Attributes settings", "Attributes settings")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	go func() {
		for {
			// systray.SetTitle(getClockTime(timezone))
			systray.SetTooltip(timezone + " timezone")
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			select {
			case <-mBrowseAttributes.ClickedCh:
				getAttributesFolder()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// Cleaning stuff here.
}

// ItoaTwoDigits time.Clock returns one digit on values, so we make sure to convert to two digits
func ItoaTwoDigits(i int) string {
	b := "0" + strconv.Itoa(i)
	return b[len(b)-2:]
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

func getAttributesFolder() {
	confFile := ReadConfig("config.toml")
	directorySelectDialog := dialog.Directory()
	directorySelectDialog.SetStartDir(confFile.Attributes.Path)
	directory, err := directorySelectDialog.Title("Find folder with attributes XML files").Browse()
	if err != nil {
		// fmt.Println("Error:", err)
	} else {
		confFile.Attributes.Path = directory
		confFile.WriteConfigParamIntoFile("config.toml")
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
