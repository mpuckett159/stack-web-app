package wshandler

import (
	"log"
	"time"

	"stack-web-app/frontend/db"
)

// RunPruner spawns the timed loop that will kick off the pruneEmptyMeetings
// goroutine.
func RunPruner() {
	// Spawn a ticker to run things every 60 seconds
	meetingPruneTicker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-meetingPruneTicker.C:
			go pruneEmptyMeetings()
		}
	}
}

// pruneEmptyMeetings will be run as a goroutine to clean up references to empty
// meetings in the database and the HubPool map so that the Go Garbage Collector
// can free up those resources (hopefully) because they are no longer referenced
func pruneEmptyMeetings() {
	// Get all active hubs in hub pool
	for hubId, hub := range HubPool {
		// Iterate through all clients and break out if an active client is found
		for _, client := range hub.clients {
			if client {
				break
			}
		}

		// If no active clients found prune meeting from hubPool and delete table from SQL
		delete(HubPool, hubId)
		log.Println("Successfully deleted meeting hub: " + hubId)
		err := db.DeleteTable(hubId)
		if err != nil {
			log.Println("Error deleting meeting table: " + err.Error())
		}
    }
}