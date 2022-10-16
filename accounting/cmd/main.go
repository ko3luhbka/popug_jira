package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/ko3luhbka/accounting/db"
	"github.com/ko3luhbka/accounting/migrations"
	"github.com/ko3luhbka/accounting/mq"
	"github.com/ko3luhbka/accounting/rest"
)

const (
	serviceName = "accounting"
)

var (
	mqCfg = &mq.Config{
		Consumer:   true,
		Producer:   false,
		ReadTopics: []string{mq.UsersCUDTopic, mq.TasksTopic},
		WriteTopic: "",
	}
)

func main() {
	log.Printf("Starting %s service", serviceName)

	conn, err := db.NewConnection()
	if err != nil {
		log.Fatal(err)
	}

	if err = db.RunMigrations(conn.DB, migrations.MigrationFiles); err != nil {
		log.Fatal(err)
	}

	accRepo := db.NewAccountRepo(conn)
	audRepo := db.NewAuditRepo(conn)

	mqClient := mq.NewMQClient(mqCfg)
	srv, err := rest.NewServer(accRepo, audRepo, mqClient)
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error)
	srv.Run(errCh)
	srv.Svc.ConsumeMsg(errCh)

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)

	select {
	case <-exitCh:
		shutdown(srv)
	case <-errCh:
		shutdown(srv)
	}
}

func shutdown(srv *rest.Server) {
	if err := srv.Shutdown(); err != nil {
		log.Println(err)
	}
	if srv.Svc.Mq.Writer != nil {
		if err := srv.Svc.Mq.Writer.Close(); err != nil {
			log.Println(err)
		}
	}
	if srv.Svc.Mq.Reader != nil {
		if err := srv.Svc.Mq.Reader.Close(); err != nil {
			log.Println(err)
		}
	}
}
