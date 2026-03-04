package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, using system environment")
	}

	down := flag.Bool("down", false, "rollback last migration")
	downAll := flag.Bool("down-all", false, "rollback all migrations")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New(
		"file://./migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}
	defer m.Close()

	if *downAll {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down-all failed: %v", err)
		}
		log.Print("All migrations rolled back successfully")
		return
	}

	if *down {
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Print("Last migration rolled back successfully")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration up failed: %v", err)
	}
	log.Print("Migrations applied successfully")
}
