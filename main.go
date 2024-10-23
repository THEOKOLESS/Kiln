package main

import (
	"fmt"
	"kiln/configuration"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] != "" {
		config, err := configuration.Init(os.Args[1])
		if err != nil {
			fmt.Println("Startup error:", err)
			os.Exit(200)
		}
		db, err := createDatabase(config.DatabaseName)
		if err != nil {
			fmt.Println("Startup error:", err)
			os.Exit(200)
		}
		defer db.Close()
		if config.AllData {
			fmt.Println("getting all delegations...")
			err = FillDbWithAllData(db)
		} else {
			fmt.Printf("getting %s latest delegations \n", config.Limit)
			err = FillDbWithLastData(db, config)

		}
		if err != nil {
			fmt.Println("Err on update Db:", err)
		}

		go func() {
			ticker := time.NewTicker(60 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				err := FillDbWithLastData(db, config)
				if err != nil {
					fmt.Println("Err on update Db:", err)
				}
			}
		}()
		http.HandleFunc("/xtz/delegations", handleRequest(db))
		fmt.Println("Server started at :8000")
		fmt.Println(http.ListenAndServe(":8000", nil))
		fmt.Println("Done!")
	} else {
		fmt.Println("You need to provide a config file.")
		os.Exit(100)
	}
	os.Exit(0)
}
