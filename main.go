package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

func main() {
	log.Printf("Start")
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/statsicon.ico"))
	cfgFile := ReadConfig("config.toml")
	attrPath := cfgFile.AttributesSettings.Path
	mBrowseAttributes := systray.AddMenuItem("Set Attributes folder", "Set Attributes folder")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	// go func() {
	// 	for {
	// 		systray.SetTooltip("AttrPath:\n" + attrPath + "\n\n ---")
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }()

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
		log.Printf("Attributes path: %s%s", attrPath, cfgFile.AttributesSettings.Filename)
		AttributeXmlOpen(attrPath + cfgFile.AttributesSettings.Filename)
	} else {
		setAttributesFolderByBrowse()
	}
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

func fileWatcher(filepath string) {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add(filepath)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func fileInfo(filename string) {
	//get file info
	fileInfo, err := os.Stat(filename)
	//handle error
	if err != nil {
		panic(err)
	}

	// print file info
	log.Printf("File info: %+v\n", fileInfo.ModTime())
}
