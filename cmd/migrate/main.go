package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"notification-system/internal/config"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	direction := flag.String("direction", "up", "migration direction: up or down")
	steps := flag.Int("steps", 0, "number of migrations to apply (0 = all)")
	flag.Parse()

	// Also support positional argument: go run cmd/migrate/main.go up
	if flag.NArg() > 0 && *direction == "up" {
		*direction = flag.Arg(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
	)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch *direction {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown direction: %s (use 'up' or 'down')\n", *direction)
		os.Exit(1)
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("no migrations to apply")
			return
		}
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Printf("migrations applied successfully (direction: %s)\n", *direction)
}
