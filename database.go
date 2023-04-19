package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func dbconnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./matchdata.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return db
}

// Check database scheme
func dbcheckscheme(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS matchkey (id INTEGER PRIMARY KEY, matchkey TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS matchdata (record_id INTEGER PRIMARY KEY, matchkey TEXT, teamsqty INTEGER, matchtype TEXT, publisher integer)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS teamplayer (record_id INTEGER PRIMARY KEY, matchrecord INTEGER, teamid INTEGER, teamisown INTEGER, profileid INTEGER, playername TEXT, playermmr INTEGER, ispartner INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS matchevent (record_id INTEGER PRIMARY KEY, matchrecord INTEGER, eventtime INTEGER, eventtype TEXT, profileid INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
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
func dbsavematchdata(db *sql.DB, m Match) {

}
