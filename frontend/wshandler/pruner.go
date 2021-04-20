package wshandler

import (
	"time"

	"stack-web-app/frontend/db"

	log "github.com/sirupsen/logrus"
)

// RunPruner spawns the timed loop that will kick off the pruneEmptyMeetings
// goroutine.
func RunPruner() {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "RunPruner",
		"module": "pruner",
	})

	// Spawn a ticker to run things every 60 seconds
	meetingPruneTicker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-meetingPruneTicker.C:
			ContextLogger.Debug("Running pruner.")
			go pruneEmptyMeetings()
		}
	}
}

// pruneEmptyMeetings will be run as a goroutine to clean up references to empty
// meetings in the database and the HubPool map so that the Go Garbage Collector
// can free up those resources (hopefully) because they are no longer referenced
func pruneEmptyMeetings() {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "pruneEmptyMeetings",
		"module": "pruner",
	})

	// Get all active hubs in hub pool
	for hubId, hub := range HubPool {
		// Iterate through all clients and break out if an active client is found
		for _, client := range hub.clients {
			if client {
				break
			}
		}
		ContextLogger.Debug("Empty meetings found, attempting to prune.")

		// If no active clients found prune meeting from hubPool and delete table from SQL
		delete(HubPool, hubId)
		ContextLogger.Debug("Successfully deleted meeting hub: " + hubId)
		err := db.DeleteTable(hubId)
		if err != nil {
			ContextLogger.Warning("Error deleting meeting table: " + err.Error())
		}
    }
	ContextLogger.Debug("Successfully pruned all empty meetings.")
}