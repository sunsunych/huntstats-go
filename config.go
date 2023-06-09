package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/sys/windows/registry"
)

type (
	AttributesSettings struct {
		Path     string
		Filename string
	}

	Activity struct {
		LastSavedKeyHash string
		Reporter         int
		SendReports      bool
		Notifications    bool
	}

	configfile struct {
		AttributesSettings AttributesSettings
		Activity           Activity
	}
)

func ReadConfig() configfile {
	var cfg configfile

	userDir, _ := os.UserConfigDir()
	configFilePath := filepath.Join(userDir, "huntstats", "config.toml")

	f := readFile(configFilePath)

	err := toml.Unmarshal(f, &cfg)
	check(err)

	isNotificationEnabled = cfg.Activity.Notifications
	isSendStatsEnabled = cfg.Activity.SendReports
	return cfg
}

func (cfg configfile) WriteConfigParamIntoFile() {
	userDir, _ := os.UserConfigDir()
	configFilePath := filepath.Join(userDir, "huntstats", "config.toml")
	cfg.AttributesSettings.Path = strings.TrimSuffix(cfg.AttributesSettings.Path, "\\attributes.xml")
	b, err := toml.Marshal(cfg)
	check(err)
	fo, err := os.Create(configFilePath)
	check(err)
	defer func() {
		if err := fo.Close(); err != nil {
			log.Printf("Config save error: %s", err)
		}
	}()
	if _, err := fo.Write(b); err != nil {
		log.Printf("Config save error: %s", err)
	}
}

func readFile(filename string) []byte {
	content, err := os.ReadFile(filename)
	check(err)
	return content
}

func GetRegSteamFolderValue() (string, error) {
	winInfo, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	check(err)
	defer winInfo.Close()

	SteamPath, _, err := winInfo.GetStringValue("InstallPath")
	LibFoldersVdf := SteamPath + "\\steamapps\\libraryfolders.vdf"
	check(err)

	attributesfolder, err := FindAttributesFolder(LibFoldersVdf)
	if err != nil {
		return "", err
	}
	return attributesfolder, nil
}
