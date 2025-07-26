/*
File: db.go
Author: trung.la
Date: 07/26/2025
Description: Database connection management for the BackEnd Monolith.
*/

package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Connections struct {
	UserDB         *sql.DB
	PomodoroDB     *sql.DB
	StatisticDB    *sql.DB
	NotificationDB *sql.DB
	TaskDB         *sql.DB
}

// NewConnections initializes all database connections.
func NewConnections(userDSN, pomodoroDSN, statisticDSN, notificationDSN, taskDSN string) (*Connections, error) {
	connections := &Connections{}

	var err error
	if connections.UserDB, err = openAndPingDB("UserDB", userDSN); err != nil {
		return nil, err
	}
	if connections.PomodoroDB, err = openAndPingDB("PomodoroDB", pomodoroDSN); err != nil {
		return nil, err
	}
	if connections.StatisticDB, err = openAndPingDB("StatisticDB", statisticDSN); err != nil {
		return nil, err
	}
	if connections.NotificationDB, err = openAndPingDB("NotificationDB", notificationDSN); err != nil {
		return nil, err
	}
	if connections.TaskDB, err = openAndPingDB("TaskDB", taskDSN); err != nil {
		return nil, err
	}

	return connections, nil
}

// openAndPingDB is a helper function to open and ping a database connection.
func openAndPingDB(name, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Error connecting to %s: %v", name, err)
		return nil, fmt.Errorf("failed to connect to %s: %w", name, err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("Error pinging %s: %v", name, err)
		return nil, fmt.Errorf("failed to ping %s: %w", name, err)
	}

	log.Printf("%s connected successfully", name)
	return db, nil
}

// Close closes all database connections.
func (c *Connections) Close() {
	closeDB("UserDB", c.UserDB)
	closeDB("PomodoroDB", c.PomodoroDB)
	closeDB("StatisticDB", c.StatisticDB)
	closeDB("NotificationDB", c.NotificationDB)
	closeDB("TaskDB", c.TaskDB)
}

// closeDB is a helper function to close a database connection.
func closeDB(name string, db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing %s: %v", name, err)
		} else {
			log.Printf("%s closed successfully", name)
		}
	}
}
