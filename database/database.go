package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"balance_checker/config"
	_ "github.com/lib/pq"
)

type DB interface {
	AddBalance(aKey string, balance float64) error
}

type Controller struct {
	db *sql.DB
}

type User struct {
	Id      float64 `json:"id"`
	User    string  `json:"user_key"`
	Balance float64 `json:"balance"`
}

//goland:noinspection ALL
const (
	createTable = `CREATE TABLE IF NOT EXISTS balance_checker (
						id 			SERIAL	PRIMARY KEY NOT NULL, 
						user_key 	VARCHAR UNIQUE 		NOT NULL,
						balance 	REAL	 			NOT NULL)`

	insertBalance = `INSERT INTO balance_checker (user_key, balance) VALUES ($1, $2)`
)

func GetController(conf *config.Config) (*Controller, *sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.DB.Host, conf.DB.Port, conf.DB.User, conf.DB.Pass, conf.DB.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("open db err: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping db err: %s", err)
	}

	c := &Controller{db: db}

	if _, err = db.Exec(createTable); err != nil {
		return nil, nil, fmt.Errorf("create table err: %s", err)
	}

	return c, db, nil
}

func (c *Controller) AddBalance(aKey string, balance float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.db.ExecContext(ctx, insertBalance, aKey, balance); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("balance add: %s %f", aKey, balance)
	return nil
}
