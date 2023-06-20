package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// List of games played by teammates
type TeamisownList struct {
	Playername  string
	Profileid   int
	Gamesplayed int
}

func dbconnection() *sql.DB {
	// var dbpath = "./data/matchdata.db"

	//NEW DB PATH
	userDir, _ := os.UserConfigDir()
	dbFolderPath := filepath.Join(userDir, "huntstats", "data")
	dbFilePatch := filepath.Join(dbFolderPath, "matchdata.db")

	checkdbexist(dbFolderPath, dbFilePatch)

	db, err := sql.Open("sqlite3", dbFilePatch)
	if err != nil {
		log.Fatal(err)
	}

	// defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	return db
}

// Check database scheme
func dbcheckscheme(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS matchkey (id INTEGER PRIMARY KEY AUTOINCREMENT, matchkey TEXT)")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS matchdata (record_id INTEGER PRIMARY KEY AUTOINCREMENT, matchkey TEXT, teamsqty INTEGER, matchtype TEXT, publisher INTEGER)")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS teamplayer (record_id INTEGER PRIMARY KEY AUTOINCREMENT, matchrecord INTEGER, teamid INTEGER, teammmr INTEGER, teamisown INTEGER, profileid INTEGER, playername TEXT, playermmr INTEGER, ispartner INTEGER)")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS matchevent (record_id INTEGER PRIMARY KEY AUTOINCREMENT, matchrecord INTEGER, eventtime INTEGER, eventtype TEXT, profileid INTEGER)")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS matchaccolade (record_id INTEGER PRIMARY KEY AUTOINCREMENT, matchrecord INTEGER, category TEXT, hits INTEGER, xp INTEGER, bounty INTEGER, weighting INTEGER, gold INTEGER)")
	checkErr(err)
}

// Read data from DB
func dbreadTemplate(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM matchevent")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var record_id int
		var matchrecord int
		var eventtime int
		var eventtype string
		var profileid int
		err = rows.Scan(&record_id, &matchrecord, &eventtime, &eventtype, &profileid)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(id, name)
	}
}

// Save matchdata
func saveNewMatchReport(m Match) {
	dbconn := dbconnection()

	//Get matchid record
	isoldmatch, matchid := getmatchid(dbconn, m.MatchKey)
	if !isoldmatch {
		log.Printf("Saving new match result")
		savematchdata(dbconn, matchid, m)
	}
}

// Get players team is own
func getPlayersTeamIsOwnList() ([]TeamisownList, error) {
	dbconn := dbconnection()

	tiol, err := getTeamIsOwnPlayers(dbconn)
	if err != nil {
		return nil, err
	}
	return tiol, nil
}

// Get playerid from latest solo game
func getPlayerIDFromLatestSolo() (int, error) {
	dbconn := dbconnection()

	plrid, err := getProfileFromLatestSoloGame(dbconn)
	if err != nil {
		return 0, err
	}
	return plrid, nil
}

func getmatchid(db *sql.DB, k string) (bool, int) {
	var id int64
	sqlStmt := "SELECT id FROM matchkey WHERE matchkey = ?"
	err := db.QueryRow(sqlStmt, k).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		res, err := db.Exec("INSERT INTO matchkey(matchkey) values(?)", k)
		if err != nil {
			log.Print(err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			log.Print(err)
		}
		return false, int(id)
	}
	return true, int(id)
}

func savematchdata(db *sql.DB, id int, m Match) {
	//Matchdata
	res, err := db.Exec("INSERT INTO matchdata(matchkey,teamsqty,matchtype) values(?,?,?)", m.MatchKey, m.TeamsQty, m.MatchType)
	if err != nil {
		log.Print(err)
	}
	matchrecord, err := res.LastInsertId()
	if err != nil {
		log.Print(err)
	}

	//Teamplayers
	for _, team := range m.Teams {
		ispartner := 0
		teamid := team.TeamID
		teamisown := team.IsOwn
		teammmr := team.TeamMMR
		for _, player := range team.Players {
			profileid := player.ProfileID
			playername := player.PlayerName
			playermmr := player.PlayerMMR
			if player.IsPartner {
				ispartner = 1
			}
			_, err := db.Exec("INSERT INTO teamplayer(matchrecord,teamid,teamisown,teammmr,profileid,playername,playermmr,ispartner) values(?,?,?,?,?,?,?,?)", matchrecord, teamid, teamisown, teammmr, profileid, playername, playermmr, ispartner)
			if err != nil {
				log.Print(err)
			}
		}
	}

	//Matchevents
	for _, matchevent := range m.Events {
		eventtime := matchevent.EventTime
		eventtype := matchevent.EventType
		profileid := matchevent.ProfileID
		_, err := db.Exec("INSERT INTO matchevent(matchrecord,eventtime,eventtype,profileid) values(?,?,?,?)", matchrecord, eventtime, eventtype, profileid)
		if err != nil {
			log.Print(err)
		}
	}

	//Accolades
	for _, matchaccolade := range m.Accolades {
		category := matchaccolade.Category
		hits := matchaccolade.Hits
		xp := matchaccolade.XP
		bounty := matchaccolade.Bounty
		weighting := matchaccolade.Weighting
		gold := matchaccolade.Gold
		_, err := db.Exec("INSERT INTO matchaccolade(matchrecord,category,hits,xp,bounty,weighting,gold) values(?,?,?,?,?,?,?)", matchrecord, category, hits, xp, bounty, weighting, gold)
		if err != nil {
			log.Print(err)
		}
	}
}

func getTeamIsOwnPlayers(db *sql.DB) ([]TeamisownList, error) {
	var PlayerIsOwnList []TeamisownList
	sqlStmt := "SELECT playername, profileid, count(profileid) as gamesplayed FROM teamplayer WHERE teamisown=1 AND ispartner=0 GROUP BY profileid ORDER BY gamesplayed DESC LIMIT 5"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tiol TeamisownList
		if err := rows.Scan(&tiol.Playername, &tiol.Profileid, &tiol.Gamesplayed); err != nil {
			return PlayerIsOwnList, err
		}
		PlayerIsOwnList = append(PlayerIsOwnList, tiol)
	}

	if err = rows.Err(); err != nil {
		return PlayerIsOwnList, err
	}

	return PlayerIsOwnList, nil
}

func getProfileFromLatestSoloGame(db *sql.DB) (int, error) {
	var plrid int
	var mrec int
	sqlStmt := "with playerdata as (select tp.matchrecord, tp.teamisown, tp.profileid, count(*) OVER (partition by tp.teamid, tp.matchrecord) AS playersinteam FROM teamplayer as tp where tp.teamisown=1) select matchrecord, profileid from playerdata where playersinteam=1 order by matchrecord limit 1"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&mrec, &plrid); err != nil {
			return 0, err
		}
		return plrid, nil
	}

	if err = rows.Err(); err != nil {
		return 0, err
	}

	return plrid, nil
}

func getPlayerNameByID(playerid int) (string, error) {
	db := dbconnection()
	var playername string
	sqlStmt := "SELECT distinct(playername) FROM teamplayer WHERE profileid=? LIMIT 1"
	err := db.QueryRow(sqlStmt, playerid).Scan(&playername)
	if err != nil {
		log.Printf("Error: ", err)
		return "", err
	}
	return playername, nil
}

func getPlayerMMRByID(playerid int) (int, error) {
	db := dbconnection()
	var playermmr int
	sqlStmt := "SELECT playermmr FROM teamplayer WHERE profileid=? ORDER BY record_id DESC LIMIT 1"
	err := db.QueryRow(sqlStmt, playerid).Scan(&playermmr)
	if err != nil {
		log.Printf("Error: ", err)
		return 0, err
	}
	return playermmr, nil
}

// DB HELPERS
func checkErr(err error) {
	if err != nil {
		log.Println("Error")
		log.Println("%s", err)
	}
}

func checkdbexist(dbpath string, dbfile string) {
	_, err := os.Stat(dbfile)
	if os.IsNotExist(err) {
		os.MkdirAll(dbpath, 0755)
		os.Create(dbfile)
	}
}
