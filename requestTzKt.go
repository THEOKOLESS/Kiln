package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Delegation struct {
	ID        int64  `json:"id"`
	Level     int    `json:"level"`
	Timestamp string `json:"timestamp"`
	Amount    int64  `json:"amount"`
	Sender    Sender `json:"sender"`
}

type Sender struct {
	Address string `json:"address"`
}

func FillDbWithAllData(db *sql.DB) error {
	lastID, err := getLastID()
	if err != nil {
		return err
	}
	fmt.Println("last id : ", lastID)
	for {
		url := fmt.Sprintf("https://api.tzkt.io/v1/operations/delegations?id.le=%d&sort.desc=id&limit=10000", lastID)
		delegations, err := fetchDelegations(url, 0)
		if err != nil {
			return err
		}
		if len(delegations) == 0 {
			fmt.Println("done gathering data")
			break
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
			}
		}
		lastID = delegations[len(delegations)-1].ID
	}
	return nil
}
func getLastID() (int64, error) {
	url := "https://api.tzkt.io/v1/operations/delegations?sort.desc=id&limit=1"
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("http get error : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("api tzkt error : %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("err while reading body : %v", err)
	}

	var delegations []Delegation
	err = json.Unmarshal(body, &delegations)
	if err != nil {
		return 0, fmt.Errorf("err while unmarshalling json : %v", err)
	}

	return delegations[0].ID, nil

}

func fetchDelegations(url string, attempt int) ([]Delegation, error) {
	fmt.Println("url : ", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get error : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if attempt == 3 {
			return nil, fmt.Errorf("api tzkt error : %d", resp.StatusCode)
		}
		fmt.Printf("api tzkt error %d \n try again in 5 sec...  \n", resp.StatusCode)
		time.Sleep(5 * time.Second)
		fetchDelegations(url, attempt+1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("err while reading body : %v", err)
	}

	var delegations []Delegation
	err = json.Unmarshal(body, &delegations)
	if err != nil {
		return nil, fmt.Errorf("err while unmarshalling json : %v", err)
	}
	fmt.Println("delegations size : ", len(delegations))
	return delegations, nil
}
