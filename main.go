package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var isNotificationEnabled = true
var isSendStatsEnabled = false

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/appicon.ico"))
	cfgFile := ReadConfig("config.toml")
	attrPath := cfgFile.AttributesSettings.Path
	mBrowseAttributes := systray.AddMenuItem("Set Attributes folder", "Set Attributes folder")
	systray.AddSeparator()
	mNotification := systray.AddMenuItemCheckbox("Notifications", "Show notifications with new results", true)
	mSync := systray.AddMenuItemCheckbox("Send stats", "Send stats to scopestats", false)
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	go func() {
		for {
			select {
			case <-mNotification.ClickedCh:
				if mNotification.Checked() {
					mNotification.Uncheck()
					isNotificationEnabled = false
					// mNotification.SetTitle("Unchecked")
				} else {
					mNotification.Check()
					isNotificationEnabled = true
					// mNotification.SetTitle("Checked")
				}
			case <-mSync.ClickedCh:
				if mSync.Checked() {
					mSync.Uncheck()
					// mNotification.SetTitle("Unchecked")
				} else {
					mSync.Check()
					// mNotification.SetTitle("Checked")
				}
			case <-mBrowseAttributes.ClickedCh:
				getAttributesFolder()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	if verifyAttributesExist(attrPath, cfgFile.AttributesSettings.Filename) {
		AttributeXmlOpen(attrPath + cfgFile.AttributesSettings.Filename)
	} else {
		getAttributesFolder()
	}

	db := dbconnection()
	dbcheckscheme(db)

	checkUpdatedAttributesFile(attrPath + cfgFile.AttributesSettings.Filename)
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
	if !verifyAttributesExist(confFile.AttributesSettings.Path, confFile.AttributesSettings.Filename) {
		directorySelectDialog.SetStartDir(GetRegSteamFolderValue())
	}
	directory, err := directorySelectDialog.Title("Find folder with attributes XML files").Browse()
	log.Printf("Selected folder: %s", directory)
	if err != nil {
		log.Println("Config set error:", err)
	} else {
		setAttributesFolderByBrowse(directory)
		// confFile.WriteConfigParamIntoFile("config.toml")
	}
	return directory
}

func setAttributesFolderByBrowse(p string) {
	confFile := ReadConfig("config.toml")
	confFile.AttributesSettings.Path = p
	confFile.WriteConfigParamIntoFile("config.toml")
}

func check(err error) {
	if err != nil {
		log.Printf("Check error: %s", err)
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
