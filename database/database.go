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

type PaginatedResults struct {
	Results     []PingResult `json:"results"`
	TotalCount  int          `json:"total_count"`
	CurrentPage int          `json:"current_page"`
	TotalPages  int          `json:"total_pages"`
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

	// Create index on timestamp for better query performance
	_, err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_timestamp ON ping_results(timestamp);")
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

func GetPingResults(page, pageSize int, timeRange string) (*PaginatedResults, error) {
	var timeFilter string
	var args []interface{}

	switch timeRange {
	case "3h":
		timeFilter = "timestamp >= datetime('now', '-3 hours')"
	case "12h":
		timeFilter = "timestamp >= datetime('now', '-12 hours')"
	case "24h":
		timeFilter = "timestamp >= datetime('now', '-24 hours')"
	case "7d":
		timeFilter = "timestamp >= datetime('now', '-7 days')"
	case "30d":
		timeFilter = "timestamp >= datetime('now', '-30 days')"
	case "2m":
		timeFilter = "timestamp >= datetime('now', '-2 months')"
	case "3m":
		timeFilter = "timestamp >= datetime('now', '-3 months')"
	default:
		timeFilter = "1=1" // No time filter
	}

	// Get total count
	var totalCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM ping_results WHERE "+timeFilter, args...).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := (totalCount + pageSize - 1) / pageSize

	// Ensure page is within bounds
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	query := "SELECT id, timestamp, latency, success FROM ping_results WHERE " + timeFilter +
		" ORDER BY timestamp DESC LIMIT ? OFFSET ?"

	rows, err := DB.Query(query, append(args, pageSize, offset)...)
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

	return &PaginatedResults{
		Results:     results,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
	}, nil
}

// GetStats returns statistics for the given time range
func GetStats(timeRange string) (map[string]interface{}, error) {
	var timeFilter string
	var args []interface{}

	switch timeRange {
	case "3h":
		timeFilter = "timestamp >= datetime('now', '-3 hours')"
	case "12h":
		timeFilter = "timestamp >= datetime('now', '-12 hours')"
	case "24h":
		timeFilter = "timestamp >= datetime('now', '-24 hours')"
	case "7d":
		timeFilter = "timestamp >= datetime('now', '-7 days')"
	case "30d":
		timeFilter = "timestamp >= datetime('now', '-30 days')"
	case "2m":
		timeFilter = "timestamp >= datetime('now', '-2 months')"
	case "3m":
		timeFilter = "timestamp >= datetime('now', '-3 months')"
	default:
		timeFilter = "1=1"
	}

	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successes,
			AVG(latency) as avg_latency,
			MIN(latency) as min_latency,
			MAX(latency) as max_latency
		FROM ping_results 
		WHERE ` + timeFilter

	var stats = make(map[string]interface{})
	var total, successes int
	var avgLatency, minLatency, maxLatency float64

	err := DB.QueryRow(query, args...).Scan(&total, &successes, &avgLatency, &minLatency, &maxLatency)
	if err != nil {
		return nil, err
	}

	stats["total"] = total
	stats["success_rate"] = float64(successes) / float64(total) * 100
	stats["avg_latency"] = avgLatency
	stats["min_latency"] = minLatency
	stats["max_latency"] = maxLatency

	return stats, nil
}

// GetGraphData returns all results within the time range for graphing
func GetGraphData(timeRange string) ([]PingResult, error) {
	var timeFilter string
	var args []interface{}

	switch timeRange {
	case "3h":
		timeFilter = "timestamp >= datetime('now', '-3 hours')"
	case "12h":
		timeFilter = "timestamp >= datetime('now', '-12 hours')"
	case "24h":
		timeFilter = "timestamp >= datetime('now', '-24 hours')"
	case "7d":
		timeFilter = "timestamp >= datetime('now', '-7 days')"
	case "30d":
		timeFilter = "timestamp >= datetime('now', '-30 days')"
	case "2m":
		timeFilter = "timestamp >= datetime('now', '-2 months')"
	case "3m":
		timeFilter = "timestamp >= datetime('now', '-3 months')"
	default:
		timeFilter = "1=1"
	}

	query := "SELECT id, timestamp, latency, success FROM ping_results WHERE " + timeFilter +
		" ORDER BY timestamp ASC"

	rows, err := DB.Query(query, args...)
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
