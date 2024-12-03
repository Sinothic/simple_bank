package main

import (
	"database/sql"
	"log"

	"github.com/Sinothic/simplebank/api"
	db "github.com/Sinothic/simplebank/db/sqlc"
	"github.com/Sinothic/simplebank/util"

	_ "github.com/lib/pq"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:root@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = ":8080"
)

var config util.Config

func init() {
	var err error
	config, err = util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
}

func main() {
	dbConnection, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(dbConnection)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
