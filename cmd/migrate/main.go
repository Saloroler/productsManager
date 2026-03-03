package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		databaseURL = flag.String("database-url", "", "PostgreSQL connection string")
		migrations  = flag.String("migrations", "file://migrations", "migrations path")
		command     = flag.String("command", "up", "migration command: up or down")
		steps       = flag.Int("steps", 0, "steps to apply for down command; 0 means all")
	)
	flag.Parse()

	if databaseURL == nil {
		log.Fatal("database-url is required")
	}

	m, err := migrate.New(*migrations, *databaseURL)
	if err != nil {
		log.Fatalf("create migrate instance: %v", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("close migration source: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("close migration db: %v", dbErr)
		}
	}()

	switch *command {
	case "up":
		err = m.Up()
	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
	default:
		log.Fatalf("unsupported command: %s", *command)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("run migrations: %v", err)
	}

	fmt.Fprintln(os.Stdout, "migrations complete")
}
