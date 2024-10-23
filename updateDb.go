package main

import (
	"database/sql"
	"fmt"
	"kiln/configuration"
)

func delegationExists(db *sql.DB, delegationID int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM data WHERE delegation_id = ? LIMIT 1)"
	err := db.QueryRow(query, delegationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("couldn't verify delegation existence : %v", err)
	}
	return exists, nil
}

func insertDelegation(db *sql.DB, delegation Delegation) error {
	insertSQL := `INSERT INTO data(delegation_id, timestamp, amount, delegator, level) VALUES (?, ?, ?, ?, ?);`

	_, err := db.Exec(insertSQL, delegation.ID, delegation.Timestamp, delegation.Amount, delegation.Sender.Address, delegation.Level)
	if err != nil {
		return fmt.Errorf("error while inserting delegation : %v", err)
	}

	return nil
}

func createDatabase(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, fmt.Errorf("fail to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("fail to connect to database: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS data (
		delegation_id INTEGER NOT NULL PRIMARY KEY,
		timestamp TEXT NOT NULL,
		amount INTEGER NOT NULL,
		delegator TEXT NOT NULL,
		level INTEGER NOT NULL
	);`

	statement, err := db.Prepare(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("fail to prepare statement: %v", err)
	}
	defer statement.Close()

	_, err = statement.Exec()
	if err != nil {
		return nil, fmt.Errorf("fail to create table: %v", err)
	}

	return db, nil
}

func FillDbWithLastData(db *sql.DB, config configuration.MainConfig) error {
	fmt.Println("updating db...")
	url := "https://api.tzkt.io/v1/operations/delegations?sort.desc=id&limit=" + config.Limit
	delegations, err := fetchDelegations(url, 0)
	if err != nil {
		return err
	}

	for _, delegation := range delegations {
		exists, err := delegationExists(db, delegation.ID)
		if err != nil {
			return err
		}

		if !exists {
			err = insertDelegation(db, delegation)
			if err != nil {
				return err
			}
		} else {
			//the rest exist in the db so we don't need to update it
			break
		}
	}
	fmt.Println("db updated")
	return nil
}
