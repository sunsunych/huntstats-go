package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbpath = "./data/matchdata.db"

func dbconnection() *sql.DB {
	checkdbexist()

	db, err := sql.Open("sqlite3", dbpath)
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
func dbsavematchdata(m Match) {
	dbconn := dbconnection()

	//Get matchid record
	isoldmatch, matchid := getmatchid(dbconn, m.MatchKey)
	if isoldmatch {
		log.Printf("Is OLD Match with MatchID: %d", matchid)
	} else {
		log.Printf("Is NEW Match with MatchID: %d", matchid)
	}

	savematchdata(dbconn, matchid, m)
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

// DB HELPERS
func checkErr(err error) {
	if err != nil {
		log.Println("Error")
		log.Println("%s", err)
	}
}

func checkdbexist() {
	_, err := os.Stat(dbpath)
	if os.IsNotExist(err) {
		os.MkdirAll("./data", 0755)
		os.Create(dbpath)
	}
}
