package main

import (
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
)

func dedup(paths ...string) {
	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Notifier error: %s", err)
	}
	defer w.Close()

	// Start listening for events.
	go dedupLoop(w)

	// Add all paths from the commandline.
	for _, p := range paths {
		err = w.Add(p)
		if err != nil {
			log.Printf("Notifier error: %s", err)
		}
		log.Printf("Starting to watch file %s", p)
	}
	<-make(chan struct{}) // Block forever
}

func dedupList(fw *fsnotify.Watcher) {
	for k, v := range fw.WatchList() {
		log.Printf("K: %d | V: %s", k, v)
	}
}

func dedupLoop(w *fsnotify.Watcher) {
	var (
		waitFor = 100 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu         sync.Mutex
		timers     = make(map[string]*time.Timer)
		printEvent = func(e fsnotify.Event) {

			readpath := strings.Split(e.String(), "\"")
			if len(readpath) > 0 {
				checkUpdatedAttributesFile(readpath[1])
			}

			mu.Lock()
			delete(timers, e.Name)
			mu.Unlock()
		}
	)

	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("ERROR: %s", err)
		case e, ok := <-w.Events:
			if !ok {
				return
			}
			if !e.Has(fsnotify.Create) && !e.Has(fsnotify.Write) {
				continue
			}

			mu.Lock()
			t, ok := timers[e.Name]
			mu.Unlock()

			if !ok {
				t = time.AfterFunc(math.MaxInt64, func() { printEvent(e) })
				t.Stop()

				mu.Lock()
				timers[e.Name] = t
				mu.Unlock()
			}

			t.Reset(waitFor)
		}
	}
}

func checkUpdatedAttributesFile(filepath string) {
	matchdata := AttributeXmlOpen(filepath)
	config := ReadConfig("config.toml")
	if matchdata.MatchKey != config.Activity.LastSavedKeyHash && len(matchdata.MatchKey) > 0 {
		if isNotificationEnabled {
			msg := buildNotificationMessageBody(matchdata)
			err := beeep.Notify("Latest match result", msg, "assets/icon.png")
			if err != nil {
				log.Printf("Notifier error: %s", err)
			}
		}
		config.Activity.LastSavedKeyHash = matchdata.MatchKey
		config.WriteConfigParamIntoFile("config.toml")
		saveNewMatchReport(matchdata)
		if config.Activity.Reporter == 0 {
			if identifyReporter() && isSendStatsEnabled == true {
				reporterid := getReporterID()
				sendMatchReport(reporterid, matchdata)
			}
		}
		if config.Activity.Reporter != 0 && isSendStatsEnabled == true {
			reporterid := getReporterID()
			sendMatchReport(reporterid, matchdata)
		}
	}
}

// exists returns whether the given file or directory exists
func checkIfFolderExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
