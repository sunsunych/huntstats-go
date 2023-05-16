package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/speps/go-hashids/v2"
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
	MatchKey  string
	TeamsQty  int
	MatchType string
	Teams     []Team
	Events    []MatchEvent
	Accolades []MatchAccolade
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

type MatchEvent struct {
	EventTime int
	EventType string
	ProfileID int
}

type MatchAccolade struct {
	Category  string `hunttag:"category"`
	Hits      int    `hunttag:"hits"`
	XP        int    `hunttag:"xp"`
	Bounty    int    `hunttag:"bounty"`
	Weighting int    `hunttag:"weighting"`
	Gold      int    `hunttag:"gold"`
}

// Interface to sort events by tmie
type ByTime []MatchEvent

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].EventTime < a[j].EventTime }

func AttributeXmlOpen(f string) Match {
	xmlFile, err := os.Open(f)
	if err != nil {
		log.Println(err)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// Attributes array
	var attributesFile Attributes
	xml.Unmarshal(byteValue, &attributesFile)
	matchdata := IterateAttributesXML(attributesFile)
	return matchdata
}

func IterateAttributesXML(attributeList Attributes) Match {
	MatchData := new(Match)
	MatchData.TeamsQty = attributeList.getTeamsAmount()
	MatchData.MatchType = attributeList.getMatchType()
	TeamsList := attributeList.getTeamsDetails(MatchData.TeamsQty)
	MatchData.Teams = TeamsList
	for _, v := range MatchData.Teams {
		teamID := v.TeamID
		playersQty := v.PlayersQty
		playersSlice := attributeList.getPlayersDetailsForTeam(teamID, playersQty)
		teamIndex := getTeamIndexByID(*MatchData, teamID)
		MatchData.Teams[teamIndex].Players = playersSlice
	}
	MatchData.MatchKey = hashMatchKey(MatchData.Teams)
	Evts := attributeList.getEventsForMatch(MatchData)
	AccoladesAmount := attributeList.getAccoladesAmount()
	Accolades := attributeList.getAccoladesDetails(AccoladesAmount)
	sort.Sort(ByTime(Evts))
	MatchData.Events = Evts
	MatchData.Accolades = Accolades
	return *MatchData
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

// Get matchtype in the match
func (a *Attributes) getMatchType() string {
	// teamsInMatch := 1
	isQuickPlay := false
	isTutorial := false
	for _, attrRecord := range a.Attr {
		if attrRecord.NameKey == "MissionBagIsQuickPlay" {
			isQuickPlay, _ = strconv.ParseBool(attrRecord.NameValue)
		}
		if attrRecord.NameKey == "MissionBagIsTutorial" {
			isTutorial, _ = strconv.ParseBool(attrRecord.NameValue)
		}
	}
	if isTutorial {
		return "tutorial"
	}
	if isQuickPlay {
		return "quickplay"
	}
	return "bountyhunt"
}

func getTeamIndexByID(md Match, teamIdx int) int {
	teamIndex := 0
	for k, v := range md.Teams {
		if teamIdx == v.TeamID {
			teamIndex = k
			return teamIndex
		}
	}
	return teamIndex
}

// Get attr value by attr key
func (a *Attributes) getValueByKey(key string) string {
	for _, attrRecord := range a.Attr {
		if attrRecord.NameKey == key {
			return attrRecord.NameValue
		}
	}
	return ""
}

// Get accolades amount in the match
func (a *Attributes) getAccoladesAmount() int {
	accoladesInMatch := 1
	for _, attrRecord := range a.Attr {
		if attrRecord.NameKey == "MissionBagNumAccolades" {
			accoladesInMatch, _ = strconv.Atoi(attrRecord.NameValue)
		}
	}
	return accoladesInMatch
}

// Get details for each team
func (a *Attributes) getTeamsDetails(teamsQty int) []Team {
	Teams := make([]Team, teamsQty)

	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionBagTeam_") {
			teamIndex, _ := getTeamIndexFromKey(attrRecord.NameKey)
			if teamIndex < teamsQty {
				Teams[teamIndex].TeamID = teamIndex
				AttrName, AttrValue := getTeamAttributeAndValue(attrRecord)
				if AttrName != "" {
					switch AttrName {
					case "mmr":
						Teams[teamIndex].TeamMMR, _ = strconv.Atoi(AttrValue)
					case "numplayers":
						Teams[teamIndex].PlayersQty, _ = strconv.Atoi(AttrValue)
					case "ownteam":
						Teams[teamIndex].IsOwn, _ = strconv.ParseBool(AttrValue)
					}
				}
			}
		}
	}
	return Teams
}

// Get details for each player
func (a *Attributes) getPlayersDetailsForTeam(teamIndex int, playersQty int) []Player {
	Players := make([]Player, playersQty)

	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionBagPlayer_") {
			teamIdx, playerIdx := getPlayerTeamAndIndexFromKey(attrRecord.NameKey)
			if teamIdx >= 0 || playerIdx >= 0 {
				if (playerIdx < playersQty) && (teamIdx == teamIndex) {
					AttrName, AttrValue := getPlayerAttributeAndValue(attrRecord)
					if AttrName != "" {
						switch AttrName {
						case "profileid":
							Players[playerIdx].ProfileID, _ = strconv.Atoi(AttrValue)
						case "mmr":
							Players[playerIdx].PlayerMMR, _ = strconv.Atoi(AttrValue)
						case "blood":
							Players[playerIdx].PlayerName = AttrValue
						case "ispartner":
							Players[playerIdx].IsPartner, _ = strconv.ParseBool(AttrValue)
						}
					}
				}
			}
		}
	}
	return Players
}

// Collect events for each player and team
func (a *Attributes) getEventsForMatch(m *Match) []MatchEvent {
	matchEvents := []MatchEvent{}
	eventRecord := MatchEvent{}
	eventattrtags := [][2]string{
		{"bountyextracted", "_extracted_bounty"},
		{"bountypickedup", "_carried_bounty"},
		{"downedbyme", "_downed"},
		{"_downedbyteammate", "_downed"},
		{"downedme", "_downed"},
		{"downedteammate", "_downed"},
		{"killedbyme", "_killed"},
		{"killedbyteammate", "_killed"},
		{"killedme", "_killed"},
		{"killedteammate", "_killed"},
	}
	for i := 0; i < m.TeamsQty-1; i++ {
		// log.Printf("Get events for Team [%d]", i)
		if len(m.Teams) > 0 {
			for pn, plr := range m.Teams[i].Players {
				playerID := plr.ProfileID
				for _, eventtag := range eventattrtags {
					keyStringTooltip := fmt.Sprintf("MissionBagPlayer_%d_%d_tooltip%s", i, pn, eventtag[0])
					tagTooltip := a.getValueByKey(keyStringTooltip)
					if tagTooltip != "" {
						matchstring := fmt.Sprintf("%s ~(\\d{1,2}):(\\d{2})", eventtag[1])
						regex := regexp.MustCompile(matchstring)
						matches := regex.FindAllStringSubmatch(tagTooltip, -1)

						for _, match := range matches {
							minutes, _ := strconv.Atoi(match[1])
							seconds, _ := strconv.Atoi(strings.TrimLeft(match[2], "0"))
							totalSeconds := minutes*60 + seconds
							eventRecord.EventTime = totalSeconds
							eventRecord.EventType = strings.TrimLeft(eventtag[0], "_")
							eventRecord.ProfileID = playerID
							matchEvents = append(matchEvents, eventRecord)
						}
					}
				}
			}
		}
	}
	return matchEvents
}

// Get accolades from mission
func (a *Attributes) getAccoladesDetails(accoladesQty int) []MatchAccolade {
	Accolades := []MatchAccolade{}
	accoladeData := MatchAccolade{}
	accoladeDataID := 0

	for _, attrRecord := range a.Attr {
		if strings.HasPrefix(attrRecord.NameKey, "MissionAccoladeEntry_") {
			accoladeIndex, _ := getAccoladeIndexFromKey(attrRecord.NameKey)
			if (accoladeDataID != accoladeIndex) && (accoladeIndex <= accoladesQty) {
				//We need to save team into Accolades slice, before assigning new accoladeIndex
				if accoladeData.Category != "fbe_bonus_generic" {
					Accolades = append(Accolades, accoladeData)
				}
				accoladeDataID = accoladeIndex
			}
			if (accoladeDataID < accoladesQty) && (accoladeDataID == accoladeIndex) {
				AttrName, AttrValue := getAccoladeAttributeAndValue(attrRecord)
				v := reflect.ValueOf(&accoladeData)

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
	return Accolades
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

// Get Accolade Index From Key
func getAccoladeIndexFromKey(key string) (int, bool) {
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

// Get Accolade Attribute From Key
func getAccoladeAttributeAndValue(rec Attr) (string, string) {
	AttrValue := ""
	KeySlice := strings.Split(rec.NameKey, "_")
	if len(KeySlice) > 2 {
		AttrName := KeySlice[2]
		if AttrName == "category" {
			AttrValue = strings.TrimPrefix(rec.NameValue, "accolade_")
		} else {
			AttrValue = rec.NameValue
		}
		return AttrName, AttrValue
	} else {
		return "", ""
	}
}

// Get Team Attribute From Key
func getPlayerAttributeAndValue(rec Attr) (string, string) {
	KeySlice := strings.Split(rec.NameKey, "_")
	if len(KeySlice) > 3 {
		AttrName := KeySlice[3]
		AttrValue := rec.NameValue
		return AttrName, AttrValue
	} else {
		return "", ""
	}
}

func hashMatchKey(teams []Team) string {
	profiles := []int{}
	for _, team := range teams {
		for _, player := range team.Players {
			profiles = append(profiles, int(player.ProfileID))
		}
	}
	hd := hashids.NewData()
	hd.Salt = HashSaltParam
	hd.MinLength = 64
	hd.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(profiles)
	return e
}

//dummy regex for time extract
// const str = '@ui_team_details_downed ~14:11~@ui_team_details_downed ~17:07';
// const regex = /(\d{1,2}):(\d{2})~/g;
// let match;
// while (match = regex.exec(str)) {
//   const minutes = match[1];
//   const seconds = match[2];
//   console.log(`Minutes: ${minutes}, Seconds: ${seconds}`);
// }

func identifyReporter() bool {
	tiol, err := getPlayersTeamIsOwnList()
	if err != nil {
		log.Printf("Error: %v", err)
	}
	if len(tiol) > 1 {
		if tiol[0].Gamesplayed > tiol[1].Gamesplayed {
			setReporterID(tiol[0].Profileid)
			return true
		}
	}
	if len(tiol) == 1 {
		if tiol[0].Gamesplayed > 2 {
			setReporterID(tiol[0].Profileid)
			return true
		}
	}
	return false
}

// get IdentytyID integer
func getReporterID() int {
	cfgFile := ReadConfig("config.toml")
	if cfgFile.Activity.Reporter != 0 {
		return cfgFile.Activity.Reporter
	} else {
		return 0
	}
}

// set IdentytyID integer
func setReporterID(reporterid int) {
	cfgFile := ReadConfig("config.toml")
	cfgFile.Activity.Reporter = reporterid
	cfgFile.WriteConfigParamIntoFile("config.toml")
}
