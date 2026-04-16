package main

import (
	"errors"
	"flag"
	"log/slog"
	"os"

	"auth_service/internal/config"
	"auth_service/internal/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	direction := flag.String("dir", "up", "migration direction: up | down")
	steps := flag.Int("steps", 0, "number of steps for up/down (0 = all)")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Error("migrations source", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.Database.URL)
	if err != nil {
		log.Error("migrate new", "error", err)
		os.Exit(1)
	}
	defer func() { _, _ = m.Close() }()

	if err := run(m, *direction, *steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("migrate", "dir", *direction, "error", err)
		os.Exit(1)
	}
	log.Info("migrations applied", "dir", *direction, "steps", *steps)
}

func run(m *migrate.Migrate, dir string, steps int) error {
	switch dir {
	case "up":
		if steps > 0 {
			return m.Steps(steps)
		}
		return m.Up()
	case "down":
		if steps > 0 {
			return m.Steps(-steps)
		}
		return m.Down()
	default:
		return errors.New("dir must be 'up' or 'down'")
	}
}
