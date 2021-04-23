package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

// User object that describes the database table columns and is used to push the info
// back to the websocket client for speaker stack rendering
type User struct {
	SpeakerPostition int16  `json:"speakerPosition"`
	SpeakerId        string `json:"speakerId"`
	Name             string `json:"name"`
}

// Start is used to start the database, that is remove any potentially existing db files
// and create the new database file. We don't care about old database contents and don't
// want it there at all so we delete before creating just to be sure.
func Start() {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "Start",
		"module":   "db",
	})

	// Delete and recreate existing sqlite file just in case
	os.Remove("sqlite-database.db")
	ContextLogger.Info("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Info("sqlite-database.db created")
}

// CreateTable is used to create a new meeting table in the database.
func CreateTable(newTableId string) (err error) {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "CreateTable",
		"module":   "db",
		"tableId":  newTableId,
	})

	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table creation SQL
	createMeetingTableSQL := "CREATE TABLE IF NOT EXISTS '" + newTableId + "' (speakerPosition INTEGER NOT NULL PRIMARY KEY, speakerId TEXT UNIQUE, name TEXT);"
	log.WithField("sqlQuery", createMeetingTableSQL).Debug("Preparing SQL query")
	statement, err := sqliteDatabase.Prepare(createMeetingTableSQL)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": createMeetingTableSQL,
			"error":    err.Error(),
		}).Error("Error preparing statement to create meeting table")
		return err
	}
	defer statement.Close()

	// Execute new table creation
	_, err = statement.Exec()
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": createMeetingTableSQL,
			"error":    err.Error(),
		}).Error("Error executing statement to create meeting table")
		return err
	}

	// Return the new table ID
	return nil
}

// DeleteTable removes a table from the SQL database to free up resources. This
// will primarily be used for removing empty meeting tables by the pruner in a
// goroutine. We will not be using vacuum here but may look into doing this later
// in a similar timed fashion and just look for the OS system free space to be
// say 1.25x the current SQL file size or something for safety.
func DeleteTable(tableId string) (err error) {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "DeleteTable",
		"module":   "db",
		"tableId":  tableId,
	})

	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table creation SQL
	deleteMeetingTableSQL := "DROP TABLE IF EXISTS '" + tableId + "';"
	log.WithField("sqlQuery", deleteMeetingTableSQL).Debug("Preparing SQL query")
	statement, err := sqliteDatabase.Prepare(deleteMeetingTableSQL)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": deleteMeetingTableSQL,
			"error":    err.Error(),
		}).Error("Error preparing statement to delete meeting table")
		return err
	}
	defer statement.Close()

	// Execute new table creation
	_, err = statement.Exec()
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": deleteMeetingTableSQL,
			"error":    err.Error(),
		}).Error("Error executing statement to delete meeting table")
		return err
	}

	// Return the new table ID
	return nil
}

// GetOnStack is the function called when a user wants to put themselves at the end of the speaker queue.
func GetOnStack(tableId string, speakerId string, name string) (err error) {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function":  "GetOnStack",
		"module":    "db",
		"tableId":   tableId,
		"speakerId": speakerId,
		"name":      name,
	})

	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table update SQL
	addUserToStackTableSQL := "INSERT INTO '" + tableId + "' (speakerId, name) VALUES (?,?);"
	log.WithField("sqlQuery", addUserToStackTableSQL).Debug("Preparing SQL query")
	statement, err := sqliteDatabase.Prepare(addUserToStackTableSQL)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": addUserToStackTableSQL,
			"error":    err.Error(),
		}).Error("Error preparing statement to get on stack.")
		return err
	}
	defer statement.Close()

	// Execute new table update
	_, err = statement.Exec(speakerId, name)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": addUserToStackTableSQL,
			"error":    err.Error(),
		}).Error("Error executing statement to get on stack.")
		return err
	}

	// Return nothing because there are no failures
	return nil
}

// GetOffStack is called when a user wants to remove themselves from the speaker queue,
// moving everyone behind them up a position.
func GetOffStack(tableId string, speakerId string) (err error) {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function":  "GetOffStack",
		"module":    "db",
		"tableId":   tableId,
		"speakerId": speakerId,
	})

	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table update SQL
	removeUserFromStackTableSQL := "DELETE FROM '" + tableId + "' WHERE speakerId=?;"
	log.WithField("sqlQuery", removeUserFromStackTableSQL).Debug("Preparing SQL query")
	statement, err := sqliteDatabase.Prepare(removeUserFromStackTableSQL)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": removeUserFromStackTableSQL,
			"error":    err.Error(),
		}).Error("Error preparing statement to get off stack")
		return err
	}
	defer statement.Close()

	// Execute new table update
	_, err = statement.Exec(speakerId)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": removeUserFromStackTableSQL,
			"error":    err.Error(),
		}).Error("Error executing statement to get off stack")
	}

	// Return nothing because there are no failures
	return nil
}

// ShowCurrent Stack is used to return the current contents of the speaker stack
// and is used on new connections and after a user has either gotten on or taken
// themselves off of the speaker stack.
func ShowCurrentStack(tableId string) (stackUsers []User, err error) {
	// Add to context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"function": "ShowCurrentStack",
		"module":   "db",
		"tableId":  tableId,
	})

	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare SELECT query
	showCurrentStackTableSQL := "SELECT speakerPosition, speakerId, name FROM '" + tableId + "';"
	log.WithField("sqlQuery", showCurrentStackTableSQL).Debug("Preparing SQL query")
	rows, err := sqliteDatabase.Query(showCurrentStackTableSQL)
	if err != nil {
		log.WithFields(log.Fields{
			"sqlQuery": showCurrentStackTableSQL,
			"error":    err.Error(),
		}).Error("Error querying meeting table")
		return nil, err
	}
	defer rows.Close()

	// Parse database rows to User object slice
	for rows.Next() {
		var stackUser User
		err := rows.Scan(&stackUser.SpeakerPostition, &stackUser.SpeakerId, &stackUser.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"sqlQuery": showCurrentStackTableSQL,
				"error":    err.Error(),
			}).Error("Error scanning query results for meeting table")
		}
		stackUsers = append(stackUsers, stackUser)
	}

	// Return current stack
	return stackUsers, nil
}
