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

// Prefix MissionBagTeam
type Team struct {
	TeamID     int `hunttag:"_id"`
	PlayersQty int `hunttag:"numplayers"`
	TeamMMR    int `hunttag:"mmr"`
	Players    []Player
	IsOwn      bool `hunttag:"ownteam"`
}

// Prefix MissionBagPlayer
type Player struct {
	ProfileID  int    `hunttag:"profileid"`
	PlayerName string `hunttag:"blood"`
	PlayerMMR  int    `hunttag:"mmr"`
	IsPartner  bool   `hunttag:"ispartner"`
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
	MatchData.Teams = TeamsList
	log.Printf("MatchData: %v", MatchData)
	for k, v := range MatchData.Teams {
		teamIndex := v.TeamID
		playersQty := v.PlayersQty
		playersSlice := attributeList.getPlayersDetailsForTeam(teamIndex, playersQty)
		log.Printf("players [%d]: %v", k, playersSlice)
	}
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
func (a *Attributes) getTeamsDetails(teamsQty int) []Team {
	Teams := []Team{}
	teamData := Team{}
	teamData.TeamID = 0

	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionBagTeam_") {
			teamIndex, _ := getTeamIndexFromKey(attrRecord.NameKey)
			if (teamData.TeamID != teamIndex) && (teamIndex <= teamsQty) {
				//We need to save team into teams slice, before assigning new teamIndex
				Teams = append(Teams, teamData)
				// log.Printf("Save Team %d with data: %v", teamData.TeamID, teamData)
				teamData.TeamID = teamIndex
			}
			if (teamData.TeamID < teamsQty) && (teamData.TeamID == teamIndex) {
				AttrName, AttrValue := getTeamAttributeAndValue(attrRecord)

				v := reflect.ValueOf(&teamData)

				for i := 0; i < reflect.Indirect(v).NumField(); i++ {
					field := reflect.Indirect(v).Field(i)
					tag := reflect.Indirect(v).Type().Field(i).Tag.Get("hunttag")
					if (tag == AttrName) && (tag != "") {
						switch reflect.Indirect(v).Field(i).Type().Name() {
						case "int":
							convertedValue, _ := strconv.Atoi(AttrValue)
							field.Set(reflect.ValueOf(convertedValue))
						case "bool":
							convertedValue, _ := strconv.ParseBool(AttrValue)
							field.Set(reflect.ValueOf(convertedValue))
						default:
							field.Set(reflect.ValueOf(AttrValue))
						}
					}
				}
			}
		}
	}
	return Teams
}

// Get details for each player
func (a *Attributes) getPlayersDetailsForTeam(teamIndex int, playersQty int) []Player {
	Players := []Player{}
	// playerData := Player{}
	playerIter := 0
	teamIter := 0
	log.Printf("Search for Team %d | %d player(s)", teamIndex, playersQty)

	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionBagPlayer_") {
			teamIdx, playerIdx := getPlayerTeamAndIndexFromKey(attrRecord.NameKey)
			if teamIdx >= 0 || playerIdx >= 0 {
				if teamIdx == teamIndex {
					if ((playerIter != playerIdx) || (teamIter != teamIndex)) && (playerIdx < playersQty) {
						//We need to save team into teams slice, before assigning new teamIndex
						// Teams = append(Teams, teamData)
						// log.Printf("[BEFORE CHANGE] PlayerIter: %d | TeamIter: %d || playerIndex: %d | teamIndex: %d", playerIter, teamIter, playerIdx, teamIdx)
						playerIter = playerIdx
						teamIter = teamIdx
						log.Printf("[AFTER  CHANGE] PlayerIter: %d | TeamIter: %d || playerIndex: %d | teamIndex: %d", playerIter, teamIter, playerIdx, teamIdx)
						// teamData.TeamID = teamIndex
					}
					// if (teamData.TeamID < teamsQty) && (teamData.TeamID == teamIndex) {
					// 	AttrName, AttrValue := getTeamAttributeAndValue(attrRecord)

					// 	v := reflect.ValueOf(&teamData)

					// 	for i := 0; i < reflect.Indirect(v).NumField(); i++ {
					// 		field := reflect.Indirect(v).Field(i)
					// 		tag := reflect.Indirect(v).Type().Field(i).Tag.Get("hunttag")
					// 		if (tag == AttrName) && (tag != "") {
					// 			switch reflect.Indirect(v).Field(i).Type().Name() {
					// 			case "int":
					// 				convertedValue, _ := strconv.Atoi(AttrValue)
					// 				field.Set(reflect.ValueOf(convertedValue))
					// 			case "bool":
					// 				convertedValue, _ := strconv.ParseBool(AttrValue)
					// 				field.Set(reflect.ValueOf(convertedValue))
					// 			default:
					// 				field.Set(reflect.ValueOf(AttrValue))
					// 			}
					// 		}
					// 	}
					// }
				}
			}
		}
	}
	return Players
}

// ITERATOR HELPERS
// Get Team Index From Key
func getTeamIndexFromKey(key string) (int, bool) {
	KeySlice := strings.Split(key, "_")
	if len(KeySlice) >= 2 {
		teamIndex, _ := strconv.Atoi(KeySlice[1])
		return teamIndex, true
	} else {
		return 0, false
	}
}

// Get player personal index and team index from key
func getPlayerTeamAndIndexFromKey(key string) (int, int) {
	KeySlice := strings.Split(key, "_")
	if len(KeySlice) >= 2 {
		teamIndex, _ := strconv.Atoi(KeySlice[1])
		playerIndex, _ := strconv.Atoi(KeySlice[2])
		return teamIndex, playerIndex
	} else {
		return -1, -1
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
