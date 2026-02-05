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
		cmd     = flag.String("command", "up", "migration command: up|down|steps|goto|version")
		dsn     = flag.String("database", "", "database DSN")
		path    = flag.String("path", "migrations", "path to migrations")
		steps   = flag.Int("steps", 1, "number of steps for steps command")
		version = flag.Int("version", 0, "target version for goto command")
	)
	flag.Parse()

	if *dsn == "" {
		fmt.Fprintln(os.Stderr, "database DSN is required")
		os.Exit(1)
	}

	m, err := migrate.New("file://"+*path, *dsn)
	if err != nil {
		log.Fatalf("migrate init error: %v", err)
	}

	switch *cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "steps":
		err = m.Steps(*steps)
	case "goto":
		err = m.Migrate(uint(*version))
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			log.Fatalf("version error: %v", verr)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return
	default:
		log.Fatalf("unknown command: %s", *cmd)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration error: %v", err)
	}
}
