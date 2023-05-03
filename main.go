package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var isNotificationEnabled = true
var isSendStatsEnabled = false

var HashSaltParam string

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/appicon.ico"))
	cfgFile := ReadConfig("config.toml")
	if HashSaltParam == "" {
		HashSaltParam = "hunt"
	}
	attrPath := cfgFile.AttributesSettings.Path

	db := dbconnection()
	dbcheckscheme(db)

	mBrowseAttributes := systray.AddMenuItem("Set Attributes folder", "Set Attributes folder")
	systray.AddSeparator()
	mNotification := systray.AddMenuItemCheckbox("Notifications", "Show notifications with new results", false)
	mSync := systray.AddMenuItemCheckbox("Send stats", "Send stats to scopestats", true)
	if cfgFile.Activity.SendReports {
		mSync.Check()
	} else {
		mSync.Uncheck()
	}
	systray.AddSeparator()
	mReportername := systray.AddMenuItem("Identify reporter", "Click to update")
	if cfgFile.Activity.Reporter != 0 {
		playername, _ := getPlayerNameByID(cfgFile.Activity.Reporter)
		playermmr, _ := getPlayerMMRByID(cfgFile.Activity.Reporter)
		titleStr := fmt.Sprintf("%s - [%d]", playername, playermmr)
		mReportername.SetTitle(titleStr)
	}
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
					isSendStatsEnabled = false
					cfgFile.Activity.SendReports = false
					cfgFile.WriteConfigParamIntoFile("config.toml")
				} else {
					mSync.Check()
					isSendStatsEnabled = true
					cfgFile.Activity.SendReports = true
					cfgFile.WriteConfigParamIntoFile("config.toml")
				}
			case <-mBrowseAttributes.ClickedCh:
				getAttributesFolder()
			case <-mReportername.ClickedCh:
				playername, _ := getPlayerNameByID(cfgFile.Activity.Reporter)
				playermmr, _ := getPlayerMMRByID(cfgFile.Activity.Reporter)
				titleStr := fmt.Sprintf("%s - [%d]", playername, playermmr)
				mReportername.SetTitle(titleStr)
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
