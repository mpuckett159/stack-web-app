package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	SpeakerPostition	int16	`json:"speakerPosition"`
	SpeakerId			string	`json:"speakerId"`
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
	createMeetingTableSQL := "CREATE TABLE IF NOT EXISTS '" + newTableId + "' (speakerPosition INTEGER NOT NULL PRIMARY KEY, speakerId TEXT UNIQUE, name TEXT)"
	log.Println("Creating new meeting table with id " + newTableId + " ...")
	log.Println("Executing " + createMeetingTableSQL)
	statement, err := sqliteDatabase.Prepare(createMeetingTableSQL)
	if err != nil {
		log.Fatal("Error preparing statement " + err.Error())
		return err
	}
	defer statement.Close()

	// Execute new table creation
	_, err = statement.Exec()
	if err != nil {
		log.Fatal("Error executing statement " + err.Error())
		return err
	}
	log.Println("Meeting table created successfully!")

	// Return the new table ID
	return nil
}

func GetOnStack(tableId string, speakerId string, name string) (err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table update SQL
	addUserToStackTableSQL := "INSERT INTO '" + tableId + "' (speakerId, name) VALUES (?,?);"
	log.Println("Adding " + name + " to stack " + tableId)
	statement, err := sqliteDatabase.Prepare(addUserToStackTableSQL)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	defer statement.Close()

	// Execute new table update
	_, err = statement.Exec(speakerId, name)
	if err != nil {
		fmt.Println("Error adding user to stack: " + err.Error())
	}

	// Return nothing because there are no failures
	return nil
}

func GetOffStack(tableId string, speakerId string) (err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare table update SQL
	removeUserFromStackTableSQL := "DELETE FROM '" + tableId + "' WHERE speakerId=?;"
	statement, err := sqliteDatabase.Prepare(removeUserFromStackTableSQL)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	defer statement.Close()

	// Execute new table update
	_, err = statement.Exec(speakerId)
	if err != nil {
		fmt.Println("Error removing user from stack: " + err.Error())
	}

	// Return nothing because there are no failures
	return nil
}

func ShowCurrentStack(tableId string) (stackUsers []User, err error) {
	// Get sqlite db connection
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()

	// Prepare SELECT query
	showCurrentStackTableSQL := "SELECT speakerPosition, speakerId, name FROM '" + tableId + "';"
	rows, err := sqliteDatabase.Query(showCurrentStackTableSQL)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	defer rows.Close()

	// Parse database rows to User object slice 
	for rows.Next() {
		var stackUser User
		err := rows.Scan(&stackUser.SpeakerPostition, &stackUser.SpeakerId, &stackUser.Name)
		if err != nil {
			fmt.Println("Error scanning DB results: " + err.Error())
		}
		stackUsers = append(stackUsers, stackUser)
	}

	// Return current stack
	return stackUsers, nil
}