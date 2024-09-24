package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var db *pgx.Conn
var redisClient *redis.Client

type Log struct {
	Message   string `json:"log_message"`
	Level     string `json:"log_level"`
	CreatedAt string `json:"created_at"`
}

func init() {
	var err error
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err = pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		slog.Error("Unable to connect to Postgres-SQL", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Postgres-SQL successfully")

	// Connect to Redis
	redisUrl := os.Getenv("REDIS_URL")
	redisOptions, err := redis.ParseURL(redisUrl)
	if err != nil {
		slog.Error("Unable to parse Redis URL", "error", err)
		os.Exit(1)
	}

	redisClient = redis.NewClient(redisOptions)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		slog.Error("Unable to connect to Redis", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Redis successfully")
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		var logEntry Log
		err := json.NewDecoder(r.Body).Decode(&logEntry)
		if err != nil {
			slog.Error("Failed to parse request body", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set the current time for created_at
		logEntry.CreatedAt = time.Now().Format(time.RFC3339)

		// Insert log entry into Postgres-SQL
		_, err = db.Exec(context.Background(), "INSERT INTO logs (log_message, log_level, created_at) VALUES ($1, $2, $3)", logEntry.Message, logEntry.Level, logEntry.CreatedAt)
		if err != nil {
			slog.Error("Failed to insert log entry into Postgres-SQL", "error", err)
			http.Error(w, "Failed to insert log", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"status": "Log entry created successfully"}`))
		if err != nil {
			slog.Error("Failed to write response", "error", err)
		}
	})
	mux.HandleFunc("GET /api/logs", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		since := query.Get("since")
		level := query.Get("level")
		limit := query.Get("limit")

		// Base query
		queryStr := "SELECT log_message, log_level, created_at FROM logs"
		var params []interface{}
		var conditions []string

		// Default to 24 hours if 'since' is not provided
		if since != "" {
			// Calculate the time range based on the 'since' value (e.g., '1h', '24h')
			duration, err := time.ParseDuration(since)
			if err != nil || duration > 24*time.Hour {
				http.Error(w, "Invalid 'since' value. Must be between 1h and 24h.", http.StatusBadRequest)
				return
			}
			fromTime := time.Now().Add(-duration).Format(time.RFC3339)
			params = append(params, fromTime)
			conditions = append(conditions, "created_at >= $1")
		}

		// Optional log level filter
		if level != "" {
			conditions = append(conditions, "log_level = $"+strconv.Itoa(len(params)+1))
			params = append(params, level)
		}

		// Combine conditions if any exist
		if len(conditions) > 0 {
			queryStr += " WHERE " + strings.Join(conditions, " AND ")
		}

		// Set limit for query
		if limit != "" {
			queryStr += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(params)+1)
			params = append(params, limit)
		} else {
			queryStr += " ORDER BY created_at DESC"
		}

		// Check if the result is in Redis
		cacheKey := fmt.Sprintf("logs:%s", r.URL.String())
		cachedResult, err := redisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			slog.Info("Cache hit. Returning logs from Redis.")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cachedResult))
			return
		}

		// Execute the query
		rows, err := db.Query(context.Background(), queryStr, params...)
		if err != nil {
			slog.Error("Failed to query logs from Postgres-SQL", "error", err)
			http.Error(w, "Failed to query logs", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var logs []Log
		for rows.Next() {
			var logEntry Log
			var createdAt time.Time

			// Scan into time.Time for the created_at field
			err := rows.Scan(&logEntry.Message, &logEntry.Level, &createdAt)
			if err != nil {
				slog.Error("Failed to scan row", "error", err)
				http.Error(w, "Failed to read logs", http.StatusInternalServerError)
				return
			}

			// Convert the timestamp back to a string
			logEntry.CreatedAt = createdAt.Format(time.RFC3339)
			logs = append(logs, logEntry)
		}

		// Convert logs to JSON
		logsJSON, err := json.Marshal(logs)
		if err != nil {
			slog.Error("Failed to marshal logs to JSON", "error", err)
			http.Error(w, "Failed to process logs", http.StatusInternalServerError)
			return
		}

		// Cache the result in Redis (for 1 hour)
		redisClient.Set(context.Background(), cacheKey, logsJSON, time.Hour)

		// Return the logs to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(logsJSON)
	})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("Server failed to start")
	}
}
