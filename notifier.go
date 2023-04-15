package main

import (
	"fmt"
	"log"
)

func buildNotificationMessageBody(m Match) string {
	msg := ""
	for _, teamSlice := range m.Teams {
		if teamSlice.IsOwn == true {
			for _, teamPlayer := range teamSlice.Players {
				msgline := fmt.Sprintf("Player: %s | MMR: %d \n", teamPlayer.PlayerName, teamPlayer.PlayerMMR)
				msg = msg + msgline
			}
		}
	}
	return msg
}

func cmdMatchResult(m Match) {
	log.Printf("Match: %s", m.MatchKey)
	log.Printf("Total %d teams", m.TeamsQty)
	for _, teamSlice := range m.Teams {
		if teamSlice.IsOwn == true {
			log.Printf("Your TeamID [%d] | Team MMR [%d]:", teamSlice.TeamID, teamSlice.TeamMMR)
			for _, teamPlayer := range teamSlice.Players {
				log.Printf("Player: %s | MMR: %d \n", teamPlayer.PlayerName, teamPlayer.PlayerMMR)
			}
		}
	}
}
