package wshandler

import (
	"time"

	"stack-web-app/frontend/db"

	log "github.com/sirupsen/logrus"
)

// pruneEmptyMeetings will be run as a goroutine to clean up references to empty
// meetings in the database and the HubPool map so that the Go Garbage Collector
// can free up those resources (hopefully) because they are no longer referenced
func PruneEmptyMeetings() {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "pruneEmptyMeetings",
		"module": "pruner",
	})

	// Get all active hubs in hub pool
	meetingPruneTicker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-meetingPruneTicker.C:
			ContextLogger.Debug("Running pruner.")
			for hubId, hub := range HubPool {
				clearHub := true
				// Iterate through all clients and break out if an active client is found
				for _, client := range hub.clients {
					if client {
						clearHub = false
						break
					}
				}

				// If no active clients found prune meeting from hubPool and delete table from SQL
				if clearHub {
					ContextLogger.Debug("Empty meeting found, attempting to prune.")
					delete(HubPool, hubId)
					ContextLogger.Debug("Successfully deleted meeting hub: " + hubId)
					err := db.DeleteTable(hubId)
					if err != nil {
						ContextLogger.Warning("Error deleting meeting table: " + err.Error())
					}
				}
			}
		}
	}
}