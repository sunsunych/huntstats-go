package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/andygrunwald/vdf"
	"github.com/sqweek/dialog"
)

type Libfolders struct {
	Folderslist map[string]interface{}
}

// var huntsteamappid = "594650"
var huntprofilespath = "\\steamapps\\common\\Hunt Showdown\\user\\profiles\\default"

func FindAttributesFolder(steamappspath string) (string, error) {
	libfolders := SearchHuntAppFolder(steamappspath)
	appspath := ""
	temppath := ""
	var err error
	// log.Printf("libfolders length: %v", reflect.ValueOf(libfolders).Len())

	if reflect.ValueOf(libfolders).Len() > 0 {
		for _, value := range libfolders {
			// libraryfolders
			for _, v := range value.(map[string]interface{}) {
				// item in library folders
				for kk, vv := range v.(map[string]interface{}) {
					if kk == "path" {
						temppath = fmt.Sprint(vv)
						foldercheck := temppath + huntprofilespath
						_, err := os.Stat(foldercheck)
						if os.IsNotExist(err) {
							log.Printf("Folder does not exist.")
						}
						appspath = foldercheck
					}
				}
			}
		}
		return appspath, nil
	}
	return "", err
}

func SearchHuntAppFolder(filepath string) map[string]interface{} {
	f, err := os.Open(filepath)
	check(err)

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		dialog.Message("Could not found installed Hunt:Showdown").Title("Error").Error()
	}

	return m
}
