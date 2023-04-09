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
	MatchID  int
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

func (t []Team) GetIndexByTeamID(tid int) int {
	for ti, teamRecord := range *t {
		if teamRecord.TeamID == tid {
			return ti
		}
	}
	return 0
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
	attributeList.getTeamsDetails(MatchData.TeamsQty)
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
	team := new(Team)
	i := -1
	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionBagTeam_") {
			teamIndex, success := getTeamIndexFromKey(attrRecord.NameKey)
			if i != teamIndex && i > -1 && i < teamsQty {
				log.Printf("==Save TeamID: %d==", team.TeamID)
				team := new(Team)
				i = teamIndex
				team.TeamID = teamIndex
			}
			if i != teamIndex || i == -1 {
				log.Printf("New team")
				team := new(Team)
				i = teamIndex
				team.TeamID = teamIndex
			}
			if success && teamIndex < (teamsQty) {
				AttrName, AttrValue := getTeamAttributeAndValue(attrRecord)
				team.TeamID = teamIndex
				v := reflect.ValueOf(team).Elem()
				for i := 0; i < v.NumField(); i++ {
					tag := v.Type().Field(i).Tag.Get("hunttag")
					if tag == AttrName {
						log.Printf("Type of %s is %v", v.Type().Field(i).Name, v.Field(i).Type().Name())
						switch v.Field(i).Type().Name() {
						case "int":
							intValue, err := strconv.ParseInt(AttrValue, 10, 64)
							if err != nil {
								log.Fatal(err)
							}
							v.Field(i).SetInt(intValue)
						case "bool":
							boolValue, err := strconv.ParseBool(AttrValue)
							if err != nil {
								log.Fatal(err)
							}
							v.Field(i).SetBool(boolValue)
						default:
							v.Field(i).SetString(AttrValue)
						}
					}
				}
				// team.SetValueByTeamHuntTag(teamIndex, AttrName, AttrValue)
				log.Printf("Team: %v", team)
				log.Printf("TeamID: %d | Parameter: %s | Value: %s \n", teamIndex, AttrName, AttrValue)
			}
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
