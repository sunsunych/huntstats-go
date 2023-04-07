package main

import (
	"fmt"
	"os"

	"github.com/andygrunwald/vdf"
)

type Libfolders struct {
	Folderslist map[string]interface{}
}

func FindAttributesFolder(steamappspath string) {
	libfolders := SearchHuntAppFolder(steamappspath)
	for key, value := range libfolders {
		fmt.Println("[", key, "] has items:")
		for _, v := range value.(map[string]interface{}) {
			// fmt.Println("\t-->", k, ":", v, "\n")
			for kk, vv := range v.(map[string]interface{}) {
				fmt.Println("\t\t-->", kk, ":", vv, "\n")
			}
		}

	}
	fmt.Println(libfolders)
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
