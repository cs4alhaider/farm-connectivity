package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type PingResult struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Latency   float64   `json:"latency"`
	Success   bool      `json:"success"`
}

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./connectivity.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS ping_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		latency REAL,
		success BOOLEAN
	);`

	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func SavePingResult(latency float64, success bool) error {
	stmt, err := DB.Prepare("INSERT INTO ping_results(latency, success) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(latency, success)
	return err
}

func GetPingResults(limit int) ([]PingResult, error) {
	rows, err := DB.Query("SELECT id, timestamp, latency, success FROM ping_results ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PingResult
	for rows.Next() {
		var result PingResult
		err := rows.Scan(&result.ID, &result.Timestamp, &result.Latency, &result.Success)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
