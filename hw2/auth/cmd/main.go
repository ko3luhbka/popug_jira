package main

import (
	"log"

	"github.com/ko3luhbka/auth/db"
	"github.com/ko3luhbka/auth/migrations"
	"github.com/ko3luhbka/auth/mq"
	"github.com/ko3luhbka/auth/rest"
)

const (
	serviceName = "auth"
)

var (
	mqCfg = &mq.Config{
		Consumer:   false,
		Producer:   true,
		ReadTopic:  "",
		WriteTopic: mq.UsersTopic,
	}
)

func main() {
	log.Printf("Starting %s service", serviceName)

	conn, err := db.NewConnection()
	if err != nil {
		log.Fatal(err)
	}

	repo := db.NewRepo(conn)

	if err = db.RunMigrations(conn.DB, migrations.MigrationFiles); err != nil {
		log.Fatal(err)
	}

	mqClient := mq.NewMQClient(mqCfg)

	srv, err := rest.NewServer(repo, mqClient)
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
