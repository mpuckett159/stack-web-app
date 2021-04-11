package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	SpeakerPostition	int16	`json:"speakerPosition"`
	Name				string	`json:"name"`
}

func Start() {
	// Delete and recreate existing sqlite file just in case
	os.Remove("sqlite-database.db")
	log.Println("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-database.db created")
}

func CreateTable(newTableId string) (err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table creation SQL
	createMeetingTableSQL := "CREATE TABLE IF NOT EXISTS '" + newTableId + "' (id INTEGER NOT NULL PRIMARY KEY,name TEXT)"
	log.Println("Creating new meeting table with id " + newTableId + " ...")
	log.Println("Executing " + createMeetingTableSQL)
	statement, err := sqliteDatabase.Prepare(createMeetingTableSQL)
	if err != nil {
		log.Fatal("Error preparing statement " + err.Error())
		return err
	}
	defer statement.Close()

	// Execute new table creation
	statement.Exec(newTableId)
	log.Println("Meeting table created successfully!")

	// Return the new table ID
	return nil
}

func GetOnStack(tableId string, name string) (err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table update SQL
	addUserToStackTableSQL := "INSERT INTO ? (name) VALUES (?);"
	log.Println("Adding " + name + " to stack " + tableId)
	statement, err := sqliteDatabase.Prepare(addUserToStackTableSQL)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	defer statement.Close()

	// Execute new table update
	statement.Exec(tableId, name)

	// Return nothing because there are no failures
	return nil
}

func GetOffStack(tableId string, name string) (err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table update SQL
	removeUserFromStackTableSQL := "DELETE FROM ? WHERE name=?;"
	log.Println("Removing " + name + " to stack " + tableId)
	statement, err := sqliteDatabase.Prepare(removeUserFromStackTableSQL)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	defer statement.Close()

	// Execute new table update
	statement.Exec(tableId, name)

	// Return nothing because there are no failures
	return nil
}

func ShowCurrentStack(tableId string) (stackUsers []User, err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare SELECT query
	showCurrentStackTableSQL := "SELECT idSpeaker, name FROM ?;"
	rows, err := sqliteDatabase.Query(showCurrentStackTableSQL, tableId)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	defer rows.Close()

	// Parse database rows to User object slice 
	for rows.Next() {
		var stackUser User
		rows.Scan(&stackUser)
		stackUsers = append(stackUsers, stackUser)
	}

	// Return current stack
	return stackUsers, nil
}