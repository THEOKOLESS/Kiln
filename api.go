package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type DelegationResponse struct {
	Data []DelegationData `json:"data"`
}

type DelegationData struct {
	Timestamp string `json:"timestamp"`
	Amount    string `json:"amount"`
	Delegator string `json:"delegator"`
	Level     string `json:"level"`
}

type Response struct {
	Description string `json:"description"`
}

func getDelegations(db *sql.DB) ([]DelegationData, error) {
	query := "SELECT timestamp, amount, delegator, level FROM data ORDER BY delegation_id DESC" //most recent first
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the delegations from the database : %v", err)
	}
	defer rows.Close()

	var delegations []DelegationData

	for rows.Next() {
		var timestamp string
		var amount int64
		var delegator string
		var level int

		err = rows.Scan(&timestamp, &amount, &delegator, &level)
		if err != nil {
			return nil, fmt.Errorf("could't read the data from the database : %v", err)
		}

		delegation := DelegationData{
			Timestamp: timestamp,
			Amount:    fmt.Sprintf("%d", amount),
			Delegator: delegator,
			Level:     fmt.Sprintf("%d", level),
		}

		delegations = append(delegations, delegation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating the rows : %v", err)
	}

	return delegations, nil
}

func handleRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RequestStatus(w, http.StatusMethodNotAllowed, "Invalid request method")
			return
		}
		delegations, err := getDelegations(db)
		response := DelegationResponse{
			Data: delegations,
		}
		if err != nil {
			RequestStatus(w, http.StatusInternalServerError, "intern server error")
			return
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Error creating JSON response:", err)
			RequestStatus(w, http.StatusInternalServerError, "intern server error")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
func RequestStatus(w http.ResponseWriter, status int, desc string) {

	response := Response{
		Description: desc,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error creating JSON response:", err)
		jsonResponse = []byte(`{Internal server error}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}
