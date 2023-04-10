package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// XML Attribute list - v37
type Attributes struct {
	XMLName     xml.Name `xml:"Attributes"`
	AttrVersion string   `xml:"Version,attr"`
	Attr        []Attr   `xml:"Attr"`
}

// XML Attribute - v37
type Attr struct {
	XMLName   xml.Name `xml:"Attr"`
	NameKey   string   `xml:"name,attr"`
	NameValue string   `xml:"value,attr"`
}

// Match details struct - v1
type Match struct {
	MatchID  uint64
	TeamsQty int
	Teams    []Team
}

func (mt *Match) GetIndexByTeamID(tid int) int {
	for ti, teamRecord := range mt.Teams {
		v := reflect.ValueOf(teamRecord)
		for i := 0; i < v.NumField(); i++ {
			tag := v.Type().Field(i).Tag.Get("hunttag")
			if tag == "_id" && mt.Teams[ti].TeamID == tid {
				return ti
			}
		}
	}
	return 0
}

type Team struct {
	TeamID     int `hunttag:"_id"`
	PlayersQty int `hunttag:"numplayers"`
	TeamMMR    int `hunttag:"mmr"`
	Players    []Player
	IsOwn      bool `hunttag:"ownteam"`
}

type Player struct {
	ProfileID  int
	PlayerName string
	PlayerMMR  int
	IsPartner  bool
}

func AttributeXmlOpen(f string) {
	xmlFile, err := os.Open(f)
	if err != nil {
		log.Println(err)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// Attributes array
	var attributesFile Attributes
	xml.Unmarshal(byteValue, &attributesFile)
	IterateAttributesXML(attributesFile)

}

func IterateAttributesXML(attributeList Attributes) {
	MatchData := new(Match)
	MatchData.TeamsQty = attributeList.getTeamsAmount()
	log.Printf("Total teams in match: %d\n", MatchData.TeamsQty)
	TeamsList := attributeList.getTeamsDetails(MatchData.TeamsQty)
	log.Printf("Teams list: %v", TeamsList)
}

// Attributes methods
// Get teams amount in the match
func (a *Attributes) getTeamsAmount() int {
	teamsInMatch := 1
	for _, attrRecord := range a.Attr {
		if attrRecord.NameKey == "MissionBagNumTeams" {
			teamsInMatch, _ = strconv.Atoi(attrRecord.NameValue)
		}
	}
	return teamsInMatch
}

// Get details for each team
func (a *Attributes) getTeamsDetails(teamsQty int) *[]Team {
	Teams := new([]Team)
	teamData := new(Team)
	lastTeamID := 0
	teamData.TeamID = lastTeamID
	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionBagTeam_") {
			teamIndex, success := getTeamIndexFromKey(attrRecord.NameKey)
			if lastTeamID != teamIndex && success {
				teamData := new(Team)
				teamData.TeamID = teamIndex
				log.Printf("Start new team ID: %d")
			} else {
				//
			}
			// log.Printf("TeamID: %d | Parameter: %s | Value: %s \n", teamIndex, AttrName, AttrValue)
		}
	}
	return Teams
}

// ITERATOR HELPERS
// Get Team Index From Key
func getTeamIndexFromKey(key string) (int, bool) {
	KeySlice := strings.Split(key, "_")
	if len(KeySlice) > 2 {
		teamIndex, _ := strconv.Atoi(KeySlice[1])
		return teamIndex, true
	} else {
		return 0, false
	}
}

// Get Team Attribute From Key
func getTeamAttributeAndValue(rec Attr) (string, string) {
	KeySlice := strings.Split(rec.NameKey, "_")
	if len(KeySlice) > 2 {
		AttrName := KeySlice[2]
		AttrValue := rec.NameValue
		return AttrName, AttrValue
	} else {
		return "", ""
	}
}

// Search in array by _id tag
func tagTest(t []Team) {
	// ValueOf returns a Value representing the run-time data
	for ti, teamRecord := range t {
		v := reflect.ValueOf(teamRecord)

		for i := 0; i < v.NumField(); i++ {
			// Get the field tag value
			tag := v.Type().Field(i).Tag.Get("hunttag")

			log.Printf("Field: %s | Tag: %s | Type: %s", v.Type().Field(i).Name, tag, v.Field(i).Type())
			log.Printf("Field index: %d | TeamID: %d", ti, t[ti].TeamID)
		}
	}
}
