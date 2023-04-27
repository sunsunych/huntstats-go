package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/getlantern/systray"
	"github.com/speps/go-hashids/v2"
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
	mBrowseAttributes := systray.AddMenuItem("Set Attributes folder", "Set Attributes folder")
	systray.AddSeparator()
	mNotification := systray.AddMenuItemCheckbox("Notifications", "Show notifications with new results", true)
	mSync := systray.AddMenuItemCheckbox("Send stats", "Send stats to scopestats", false)
	systray.AddSeparator()
	mTestRequest := systray.AddMenuItemCheckbox("Send test request", "Send test request", false)
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
			case <-mTestRequest.ClickedCh:
				sendTestRequest()
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

func sendTestRequest() {
	log.Printf("I will send test request to scopestats")
	log.Printf("With hash salt: %s", HashSaltParam)

	reporter := make([]int, 0)
	reporter = append(reporter, 55834722896)

	hd := hashids.NewData()
	hd.Salt = HashSaltParam
	hd.MinLength = 64
	hd.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(reporter)

	url := "http://127.0.0.1:3000/"
	contentType := "application/json"
	data := []byte(`{"name": "Test User", "email": "test@example.com"}`)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("X-Reporter", e)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}
