package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/icon_32.ico"))
	cfgFile := ReadConfig("config.toml")
	attrPath := cfgFile.AttributesSettings.Path
	mBrowseAttributes := systray.AddMenuItem("Set Attributes folder", "Set Attributes folder")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	go func() {
		for {
			select {
			case <-mBrowseAttributes.ClickedCh:
				setAttributesFolderByBrowse()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	if verifyAttributesExist(attrPath, cfgFile.AttributesSettings.Filename) {
		AttributeXmlOpen(attrPath + cfgFile.AttributesSettings.Filename)
	} else {
		setAttributesFolderByBrowse()
	}

	watchPath := attrPath + cfgFile.AttributesSettings.Filename
	dedup(watchPath)
}

func onExit() {
	// Cleaning stuff here.
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		log.Print(err)
	}
	return b
}

func getAttributesFolder() string {
	confFile := ReadConfig("config.toml")
	directorySelectDialog := dialog.Directory()
	if verifyAttributesExist(confFile.AttributesSettings.Path, confFile.AttributesSettings.Filename) {
		directorySelectDialog.SetStartDir(confFile.AttributesSettings.Path)
	} else {
		directorySelectDialog.SetStartDir(GetRegSteamFolderValue())
	}
	directory, err := directorySelectDialog.Title("Find folder with attributes XML files").Browse()
	if err != nil {
		// fmt.Println("Error:", err)
	} else {
		// fileInfo(confFile.AttributesSettings.Path + "\\attributes.xml")
		confFile.AttributesSettings.Path = directory
		confFile.WriteConfigParamIntoFile("config.toml")
	}
	return directory
}

func setAttributesFolderByBrowse() {
	confFile := ReadConfig("config.toml")
	directorySelectDialog := dialog.Directory()
	directorySelectDialog.SetStartDir(confFile.AttributesSettings.Path)
	directory, err := directorySelectDialog.Title("Find folder with attributes XML files").Browse()
	if err != nil {
		// fmt.Println("Error:", err)
	} else {
		// fileInfo(confFile.AttributesSettings.Path + "\\attributes.xml")
		confFile.AttributesSettings.Path = directory
		confFile.WriteConfigParamIntoFile("config.toml")
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func verifyAttributesExist(folderpath string, filename string) bool {
	log.Printf("Check attributes at %s%s", folderpath, filename)
	fullPath := folderpath + filename
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		log.Printf("File %s is not found at %s", filename, folderpath)
		return false
	}
	return true
}

func dedup(paths ...string) {
	if len(paths) < 1 {
		// exit("must specify at least one path to watch")
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		// exit("creating a new watcher: %s", err)
	}
	defer w.Close()

	// Start listening for events.
	go dedupLoop(w)

	// Add all paths from the commandline.
	for _, p := range paths {
		err = w.Add(p)
		if err != nil {
			// exit("%q: %s", p, err)
		}
	}

	// log.Printf("ready; press ^C to exit")
	<-make(chan struct{}) // Block forever
}

func dedupLoop(w *fsnotify.Watcher) {
	var (
		// Wait 100ms for new events; each new event resets the timer.
		waitFor = 100 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)

		// Callback we run.
		printEvent = func(e fsnotify.Event) {
			// log.Printf(e.String())
			readpath := strings.Split(e.String(), "\"")
			if len(readpath) > 0 {
				log.Printf("Attr UPD path: %s", readpath[1])
				matchdata := AttributeXmlOpen(readpath[1])
				msg := buildNotificationMessageBody(matchdata)
				err := beeep.Notify("Lates match result", msg, "assets/icon.png")
				if err != nil {
					// panic(err)
				}
			}

			// Don't need to remove the timer if you don't have a lot of files.
			mu.Lock()
			delete(timers, e.Name)
			mu.Unlock()
		}
	)

	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Printf("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// We just want to watch for file creation, so ignore everything
			// outside of Create and Write.
			if !e.Has(fsnotify.Create) && !e.Has(fsnotify.Write) {
				continue
			}

			// Get timer.
			mu.Lock()
			t, ok := timers[e.Name]
			mu.Unlock()

			// No timer yet, so create one.
			if !ok {
				t = time.AfterFunc(math.MaxInt64, func() { printEvent(e) })
				t.Stop()

				mu.Lock()
				timers[e.Name] = t
				mu.Unlock()
			}

			// Reset the timer for this path, so it will start from 100ms again.
			t.Reset(waitFor)
		}
	}
}

func buildNotificationMessageBody(m Match) string {
	msg := ""
	// log.Printf("[MY TEAM]")
	for _, teamSlice := range m.Teams {
		if teamSlice.IsOwn == true {
			for _, teamPlayer := range teamSlice.Players {
				msgline := fmt.Sprintf("Player: %s | MMR: %d \n", teamPlayer.PlayerName, teamPlayer.PlayerMMR)
				msg = msg + msgline
			}
		}
	}
	return msg
}
