package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var isNotificationEnabled = false
var isSendStatsEnabled = false
var isDebugParam = "false"

var HashSaltParam string
var ReportServer string

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/appicon.ico"))
	cfgFile := ReadConfig("config.toml")

	// Logfile setup
	f, err := os.OpenFile("events.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Check isDebug param
	isDebug, _ := strconv.ParseBool(isDebugParam)
	if isDebug {
		log.Printf("Started as debug build")
		ReportServer = "http://127.0.0.1:3000"
	} else {
		ReportServer = "https://api.scopestats.com"
	}
	log.Printf("DEBUG MODE IS: %v", isDebug)

	// Hash parameter setup
	if HashSaltParam == "" {
		HashSaltParam = "hunt"
	}

	// Set file attributes.xml path
	attrPath := fmt.Sprintf("%s%s", cfgFile.AttributesSettings.Path, cfgFile.AttributesSettings.Filename)
	isAttributesPathValid := false
	// Steps to check if attributes file is correct
	// 1. Validate config path
	if !isAttributesPathValid {
		log.Printf("Start to check if attributes.xml path is valid in config")
		isAttributesPathValid = verifyAttributesExist(attrPath)
		if isAttributesPathValid {
			attrPath = attrPath
		}
		log.Printf("Config path is %t", isAttributesPathValid)
	}
	// 2. If it not ok with config path - check regedit
	if !isAttributesPathValid {
		log.Printf("Start to check if attributes.xml path is valid in windows registry")
		steamfolder := GetRegSteamFolderValue()
		log.Printf("App folder in windows registry: %s", steamfolder)
		isAttributesPathValid = verifyAttributesExist(steamfolder)
		if isAttributesPathValid {
			attrPath = steamfolder
		}
		log.Printf("Windows registry path is %t", isAttributesPathValid)
		// Add save param into config
	}
	// 3. If hunt is not installed - show dialog with folder selection
	if !isAttributesPathValid {
		log.Printf("Start to check if attributes.xml path from directory selection window")
		directorySelectDialogPath, err := dialog.Directory().SetStartDir(attrPath).Title("Find folder with attributes XML files").Browse()
		if err != nil {
			log.Print(err)
		}
		isAttributesPathValid = verifyAttributesExist(directorySelectDialogPath)
		if isAttributesPathValid {
			attrPath = directorySelectDialogPath
		}
	}

	cfgFile.AttributesSettings.Path = attrPath
	cfgFile.WriteConfigParamIntoFile("config.toml")

	checkUpdatedAttributesFile(attrPath + cfgFile.AttributesSettings.Filename)
	watchPath := attrPath + cfgFile.AttributesSettings.Filename
	// 4. On updated config attribute path add it into watch path
	dedup(watchPath)

	// Database connection init
	db := dbconnection()
	dbcheckscheme(db)

	// Tray icon and menu setup
	mBrowseAttributes := systray.AddMenuItem("Set Attributes folder", "Set Attributes folder")
	systray.AddSeparator()
	mNotification := systray.AddMenuItemCheckbox("Notifications", "Show notifications with new results", true)
	if cfgFile.Activity.Notifications {
		mNotification.Check()
	} else {
		mNotification.Uncheck()
	}
	mSync := systray.AddMenuItemCheckbox("Send stats", "Send stats to scopestats", true)
	if cfgFile.Activity.SendReports {
		mSync.Check()
	} else {
		mSync.Uncheck()
	}
	mReportername := systray.AddMenuItem("- Unkown reporter -", "Click to update")
	if cfgFile.Activity.Reporter != 0 {
		playername, err := getPlayerNameByID(cfgFile.Activity.Reporter)
		if err != nil {
			mReportername.Hide()
		}
		playermmr, err := getPlayerMMRByID(cfgFile.Activity.Reporter)
		if err != nil {
			mReportername.Hide()
		}
		if playername != "" && playermmr > 0 {
			titleStr := fmt.Sprintf("%s - [%d]", playername, playermmr)
			mReportername.SetTitle(titleStr)
			mReportername.Show()
		}
	}
	if cfgFile.Activity.Reporter == 0 {
		mReportername.Hide()
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
					cfgFile.Activity.Notifications = false
					cfgFile.WriteConfigParamIntoFile("config.toml")
				} else {
					mNotification.Check()
					isNotificationEnabled = true
					cfgFile.Activity.Notifications = true
					cfgFile.WriteConfigParamIntoFile("config.toml")
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
				if cfgFile.Activity.Reporter != 0 {
					playername, err := getPlayerNameByID(cfgFile.Activity.Reporter)
					if err != nil {
						mReportername.Disable()
					}
					playermmr, err := getPlayerMMRByID(cfgFile.Activity.Reporter)
					if err != nil {
						mReportername.Disable()
					}
					if !mReportername.Disabled() {
						titleStr := fmt.Sprintf("%s - [%d]", playername, playermmr)
						mReportername.SetTitle(titleStr)
					}
				}
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

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		log.Print(err)
	}
	return b
}

func getAttributesFolder() {
	confFile := ReadConfig("config.toml")
	filepath := fmt.Sprintf("%s%s", confFile.AttributesSettings.Path, confFile.AttributesSettings.Filename)
	directorySelectDialog := dialog.Directory()
	if !verifyAttributesExist(filepath) {
		directorySelectDialog.SetStartDir(GetRegSteamFolderValue())
	}
	directory, err := directorySelectDialog.Title("Find folder with attributes XML files").Browse()
	if err != nil {
		log.Println("Config set error:", err)
	} else {
		setAttributesFolderByBrowse(directory)
		// confFile.WriteConfigParamIntoFile("config.toml")
	}
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

func verifyAttributesExist(folderpath string) bool {
	_, err := os.Stat(folderpath)
	if os.IsNotExist(err) {
		log.Printf("File attributes.xml is not found at %s", folderpath)
		return false
	}
	return true
}
