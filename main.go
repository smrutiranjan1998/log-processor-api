package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	var err error
	connStr := "postgres://admin:password@localhost:5432/logdb?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
}

func processLogFile(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("logs/processed_logs.txt")
	if err != nil {
		http.Error(w, "Error reading log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}

		logTime := time.Now()
		logLevel := parts[0]
		message := parts[2]

		_, err := db.Exec(`INSERT INTO logs (log_time, log_level, message) VALUES ($1, $2, $3)`, logTime, logLevel, message)
		if err != nil {
			fmt.Printf("Error inserting log: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		http.Error(w, "Error reading log file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Log file processed and data inserted into the database"))
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT log_time, log_level, message FROM logs")
	if err != nil {
		http.Error(w, "Error retrieving logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var logTime time.Time
		var logLevel, message string
		rows.Scan(&logTime, &logLevel, &message)

		logs = append(logs, map[string]interface{}{
			"log_time":  logTime,
			"log_level": logLevel,
			"message":   message,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/process", processLogFile)
	http.HandleFunc("/logs", getLogs)

	log.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
