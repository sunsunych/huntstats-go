package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andygrunwald/vdf"
)

type Libfolders struct {
	Folderslist map[string]interface{}
}

// var huntsteamappid = "594650"
var huntprofilespath = "\\steamapps\\common\\Hunt Showdown\\user\\profiles\\default"

func FindAttributesFolder(steamappspath string) string {
	libfolders := SearchHuntAppFolder(steamappspath)
	appspath := ""
	temppath := ""

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
	return appspath
}

func SearchHuntAppFolder(filepath string) map[string]interface{} {
	f, err := os.Open(filepath)
	check(err)

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		panic(err)
	}

	return m
}
