/*
File: db.go
Author: trung.la
Date: 07/26/2025
Description: Database connection management for the BackEnd Monolith.
*/

package database

import (
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

type Connections struct {
	UserDB         *sql.DB
	PomodoroDB     *sql.DB
	StatisticDB    *sql.DB
	NotificationDB *sql.DB
	TaskDB         *sql.DB
}

func NewConnections(userDSN, pomodoroDSN, statisticDSN, notificationDSN, taskDSN string) (*Connections, error) {
	userDB, err := sql.Open("postgres", userDSN)
	if err != nil {
		log.Printf("Error connecting to UserDB: %v", err)
		return nil, err
	}

	if err := userDB.Ping(); err != nil {
		log.Printf("Error pinging UserDB: %v", err)
		return nil, err
	}

	pomodoroDB, err := sql.Open("postgres", pomodoroDSN)
	if err != nil {
		log.Printf("Error connecting to PomodoroDB: %v", err)
		return nil, err
	}

	if err := pomodoroDB.Ping(); err != nil {
		log.Printf("Error pinging PomodoroDB: %v", err)
		return nil, err
	}

	statisticDB, err := sql.Open("postgres", statisticDSN)
	if err != nil {
		log.Printf("Error connecting to StatisticsDB: %v", err)
		return nil, err
	}

	if err := statisticDB.Ping(); err != nil {
		log.Printf("Error pinging StatisticsDB: %v", err)
		return nil, err
	}

	notificationDB, err := sql.Open("postgres", notificationDSN)
	if err != nil {
		log.Printf("Error connecting to NotificationDB: %v", err)
		return nil, err
	}

	if err := notificationDB.Ping(); err != nil {
		log.Printf("Error pinging NotificationDB: %v", err)
		return nil, err
	}

	taskDB, err := sql.Open("postgres", taskDSN)
	if err != nil {
		log.Printf("Error connecting to TaskManagementDB: %v", err)
		return nil, err
	}

	if err := taskDB.Ping(); err != nil {
		log.Printf("Error pinging TaskManagementDB: %v", err)
		return nil, err
	}

	return &Connections{
		UserDB:         userDB,
		PomodoroDB:     pomodoroDB,
		StatisticDB:    statisticDB,
		NotificationDB: notificationDB,
		TaskDB:         taskDB,
	}, nil
}

func (c *Connections) Close() {
	if c.UserDB != nil {
		if err := c.UserDB.Close(); err != nil {
			log.Printf("Error closing UserDB: %v", err)
		}
	}
	if c.PomodoroDB != nil {
		if err := c.PomodoroDB.Close(); err != nil {
			log.Printf("Error closing PomodoroDB: %v", err)
		}
	}
	if c.StatisticDB != nil {
		if err := c.StatisticDB.Close(); err != nil {
			log.Printf("Error closing StatisticsDB: %v", err)
		}
	}
	if c.NotificationDB != nil {
		if err := c.NotificationDB.Close(); err != nil {
			log.Printf("Error closing NotificationDB: %v", err)
		}
	}
	if c.TaskDB != nil {
		if err := c.TaskDB.Close(); err != nil {
			log.Printf("Error closing TaskDB: %v", err)
		}
	}
}
