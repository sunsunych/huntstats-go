package main

import (
	"log"
	"os"

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
	}

	configfile struct {
		AttributesSettings AttributesSettings
		Activity           Activity
	}
)

func ReadConfig(filename string) configfile {
	var cfg configfile

	f := readFile(filename)

	err := toml.Unmarshal(f, &cfg)
	check(err)
	return cfg
}

func (cfg configfile) WriteConfigParamIntoFile(filename string) {
	b, err := toml.Marshal(cfg)
	check(err)
	fo, err := os.Create(filename)
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

func GetRegSteamFolderValue() string {
	winInfo, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	check(err)
	defer winInfo.Close()

	SteamPath, _, err := winInfo.GetStringValue("InstallPath")
	LibFoldersVdf := SteamPath + "\\steamapps\\libraryfolders.vdf"
	check(err)

	attributesfolder := FindAttributesFolder(LibFoldersVdf)
	return attributesfolder
}
