package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var appVersion = "0.1.6"

var isNotificationEnabled = false
var isSendStatsEnabled = false
var isDebugParam = "false"

var HashSaltParam string
var ReportServer string
var ProfileServer string

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex = kernel32.NewProc("CreateMutexW")
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	settingsMigrationCheck()
	systray.SetIcon(getIcon("assets/appicon.ico"))
	cfgFile := ReadConfig()

	// Logfile setup
	userDir, _ := os.UserConfigDir()
	logFilePath := filepath.Join(userDir, "huntstats", "events.log")
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// mutexName starting with "Global\" will work across all user sessions
	_, err = CreateMutex("HUNTSTATS")
	if err != nil {
		log.Printf("copy app is already running: %v", err)
		dialog.Message("Application is already running").Title("Application is running").Error()
		systray.Quit()
	}

	// Check isDebug param
	isDebug, _ := strconv.ParseBool(isDebugParam)
	if isDebug {
		ReportServer = "http://127.0.0.1:3000"
		ProfileServer = "http://localhost:5173"
		log.Printf("[DEBUG MODE ENABLED]")
	} else {
		ReportServer = "https://api.scopestats.com"
		ProfileServer = "https://scopestats.com"
	}

	// Hash parameter setup
	if HashSaltParam == "" {
		HashSaltParam = "hunt"
	}

	// Set file attributes.xml path
	attrPath := fmt.Sprintf("%s%s", cfgFile.AttributesSettings.Path, cfgFile.AttributesSettings.Filename)
	isAttributesPathValid := false

	// Steps to check if attributes file is correct
	// 1. Validate config path
	//Start to check if attributes.xml path is valid in config
	if !isAttributesPathValid {
		isAttributesPathValid = verifyAttributesExist(attrPath)
	}
	// 2. If it not ok with config path - check regedit
	//Start to check if attributes.xml path is valid in windows registry
	if !isAttributesPathValid {
		steamfolder, err := GetRegSteamFolderValue()
		if err != nil {
			isAttributesPathValid = false
		}
		isAttributesPathValid = verifyAttributesExist(steamfolder)
		if isAttributesPathValid {
			attrPath = steamfolder
		}
		log.Printf("Windows registry path is %t", isAttributesPathValid)
	}
	// 3. If hunt is not installed - show dialog with folder selection
	// Start to check if attributes.xml path from directory selection window
	if !isAttributesPathValid {
		directorySelectDialogPath, err := dialog.Directory().SetStartDir(attrPath).Title("Find folder with attributes XML files").Browse()
		if err != nil {
			log.Print(err)
		}
		directorySelectedFilePath := fmt.Sprintf("%s%s", directorySelectDialogPath, cfgFile.AttributesSettings.Filename)
		isAttributesPathValid = verifyAttributesExist(directorySelectedFilePath)
		if isAttributesPathValid {
			attrPath = directorySelectedFilePath
		}
	}

	// Database connection init
	db := dbconnection()
	dbcheckscheme(db)

	// Tray icon and menu setup
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
	mOnlineProfile := systray.AddMenuItem("My profile on Scopestats", "Open my scopestats profile")
	if cfgFile.Activity.Reporter != 0 {
		playername, err := getPlayerNameByID(cfgFile.Activity.Reporter)
		if err != nil {
			mReportername.Hide()
			mOnlineProfile.Hide()
		}
		playermmr, err := getPlayerMMRByID(cfgFile.Activity.Reporter)
		if err != nil {
			mReportername.Hide()
			mOnlineProfile.Hide()
		}
		if playername != "" && playermmr > 0 {
			titleStr := fmt.Sprintf("%s - [%d]", playername, playermmr)
			mReportername.SetTitle(titleStr)
			mReportername.Show()
			mOnlineProfile.Show()
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
					cfgFile = ReadConfig()
					isNotificationEnabled = false
					cfgFile.Activity.Notifications = false
					cfgFile.WriteConfigParamIntoFile()
				} else {
					mNotification.Check()
					cfgFile = ReadConfig()
					isNotificationEnabled = true
					cfgFile.Activity.Notifications = true
					cfgFile.WriteConfigParamIntoFile()
				}
			case <-mSync.ClickedCh:
				if mSync.Checked() {
					mSync.Uncheck()
					cfgFile = ReadConfig()
					isSendStatsEnabled = false
					cfgFile.Activity.SendReports = false
					cfgFile.WriteConfigParamIntoFile()
				} else {
					mSync.Check()
					cfgFile = ReadConfig()
					isSendStatsEnabled = true
					cfgFile.Activity.SendReports = true
					cfgFile.WriteConfigParamIntoFile()
				}
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
			case <-mOnlineProfile.ClickedCh:
				if cfgFile.Activity.Reporter != 0 {
					profileLink := buildReporterProfileLink(cfgFile.Activity.Reporter)
					openProfile(profileLink)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	if isAttributesPathValid {
		cfgFile.AttributesSettings.Path = attrPath
		cfgFile.WriteConfigParamIntoFile()
		checkUpdatedAttributesFile(attrPath)
		dedup(attrPath)
	} else {
		log.Printf("Unable to locate attributes.xml file")
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

func getAttributesFolder(startpath string) {
	log.Printf("Init new attributes folder to watch. Start exploring from: %s", startpath)
	confFile := ReadConfig()
	// filepath := fmt.Sprintf("%s%s", confFile.AttributesSettings.Path, confFile.AttributesSettings.Filename)
	directorySelectDialog := dialog.Directory().SetStartDir(startpath)
	directory, err := directorySelectDialog.Title("Find folder with attributes XML files").Browse()
	if err != nil {
		log.Println("Config set error:", err)
	} else {
		log.Printf("Selected directory: %s", directory)
		setAttributesFolderByBrowse(directory)
		confFile.WriteConfigParamIntoFile()
	}
}

func setAttributesFolderByBrowse(p string) {
	log.Printf("Path %s will be saved as folder with attributes.xml", p)
	confFile := ReadConfig()
	confFile.AttributesSettings.Path = p
	confFile.WriteConfigParamIntoFile()
}

func check(err error) {
	if err != nil {
		log.Printf("Check error: %s", err)
	}
}

func verifyAttributesExist(filepath string) bool {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		log.Printf("File attributes.xml is not found at %s", filepath)
		return false
	}
	return true
}

func openProfile(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func appIsRunning(appname string) bool {
	cmd := exec.Command("TASKLIST", "/FI", fmt.Sprintf("IMAGENAME eq %s", appname))
	result, err := cmd.Output()
	if err != nil {
		return false
	}
	return !bytes.Contains(result, []byte("No tasks are running"))
}

func CreateMutex(name string) (uintptr, error) {
	ret, _, err := procCreateMutex.Call(
		0,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	switch int(err.(syscall.Errno)) {
	case 0:
		return ret, nil
	default:
		return ret, err
	}
}

func settingsMigrationCheck() {
	userDir, _ := os.UserConfigDir()
	configFilePath := filepath.Join("makefile")
	dbFilePath := filepath.Join("data", "matchdata.db")

	filepaths := [2]string{configFilePath, dbFilePath}
	for _, fp := range filepaths {
		newFullPath := filepath.Join(userDir, "huntstats", fp)
		_, err := os.Stat(fp)
		if os.IsNotExist(err) {
			if isDebugParam == "true" {
				log.Printf("File %s is not found or is already migrated", fp)
			}
		} else {
			_, err = os.Stat(newFullPath)
			if os.IsNotExist(err) {
				userDirFolder := filepath.Join(userDir, "huntstats")
				os.MkdirAll(userDirFolder, 0755)
				dbDirFolder := filepath.Join(userDir, "huntstats", "data")
				os.MkdirAll(dbDirFolder, 0755)
			}
			err = MoveFile(fp, newFullPath)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func MoveFile(source, destination string) (err error) {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destination, flag, perm)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destination)
		return err
	}
	err = dst.Close()
	if err != nil {
		return err
	}
	err = src.Close()
	if err != nil {
		return err
	}
	err = os.Remove(source)
	if err != nil {
		return err
	}
	return nil
}
