package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host string `yaml:"host"` // Хост
	Port string `yaml:"port"` // Порт
	User string `yaml:"user"` // Пользователь
	Pass string `yaml:"pass"` // Пароль
	Name string `yaml:"name"` // Название
}

type DB interface {
	AddBalance(aKey string, balance float64) error
}

type Store struct {
	db *sql.DB
}

type User struct {
	Id      string  `json:"id"`
	Balance float64 `json:"balance"`
}

//goland:noinspection ALL
const (
	createTable = `CREATE TABLE IF NOT EXISTS balance_checker (
						id 			VARCHAR	PRIMARY KEY NOT NULL, 
						balance 	REAL	 			NOT NULL)`

	insertBalance = `INSERT INTO balance_checker (id, balance) VALUES ($1, $2) 
						ON CONFLICT(id) DO UPDATE SET balance = $2`
)

func NewStore(ctx context.Context, conf *Config) (*Store, *sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Pass, conf.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("open db err: %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping db err: %s", err)
	}

	if _, err = db.Exec(createTable); err != nil {
		return nil, nil, fmt.Errorf("create table err: %s", err)
	}

	return &Store{db: db}, db, nil
}

func (c *Store) AddBalance(aKey string, balance float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.db.ExecContext(ctx, insertBalance, aKey, balance); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("balance add/update: %s %f", aKey, balance)
	return nil
}
