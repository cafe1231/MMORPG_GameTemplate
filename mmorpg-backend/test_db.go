package main

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/lib/pq"
)

func main() {
	// Try different connection strings
	connStrs := []string{
		"postgres://dev:dev@localhost:5432/mmorpg?sslmode=disable",
		"postgres://dev:dev@127.0.0.1:5432/mmorpg?sslmode=disable",
		"postgres://dev:dev@host.docker.internal:5432/mmorpg?sslmode=disable",
		"host=localhost port=5432 user=dev password=dev dbname=mmorpg sslmode=disable",
		"host=127.0.0.1 port=5432 user=dev password=dev dbname=mmorpg sslmode=disable",
	}
	
	for i, connStr := range connStrs {
		fmt.Printf("\n=== Testing connection string %d ===\n", i+1)
		fmt.Printf("Connection string: %s\n", connStr)
		
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open database: %v", err)
			continue
		}
		
		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database: %v", err)
			db.Close()
			continue
		}
		
		fmt.Println("SUCCESS! Connected to database")
		
		// Test a simple query
		var result int
		err = db.QueryRow("SELECT 1").Scan(&result)
		if err != nil {
			log.Printf("Failed to execute query: %v", err)
		} else {
			fmt.Printf("Query result: %d\n", result)
		}
		
		db.Close()
	}
}