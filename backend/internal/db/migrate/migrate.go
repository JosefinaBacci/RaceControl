package migrate

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/racecontrol/backend/migrations"
)

// Migrator applies the SQL migrations embedded in the binary.
type Migrator struct {
	engine *migrate.Migrate
	db     *sql.DB
}

// New builds a Migrator over the embedded migrations.
func New(databaseURL string) (*Migrator, error) {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database for migrations: %w", err)
	}

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("create migration driver: %w", err), db.Close())
	}

	engine, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("create migrator: %w", err), db.Close())
	}

	return &Migrator{engine: engine, db: db}, nil
}

// Up applies every pending migration and reports the resulting version.
func (m *Migrator) Up() (uint, error) {
	if err := m.engine.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}
	return m.Version()
}

// Down rolls back the given number of migrations, or every one of them when steps
// is not positive.
func (m *Migrator) Down(steps int) (uint, error) {
	var err error
	if steps > 0 {
		err = m.engine.Steps(-steps)
	} else {
		err = m.engine.Down()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, fmt.Errorf("roll back migrations: %w", err)
	}
	return m.Version()
}

// Version reports the applied migration version and whether it is dirty.
func (m *Migrator) Version() (uint, error) {
	version, dirty, err := m.engine.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read migration version: %w", err)
	}
	if dirty {
		return version, fmt.Errorf("migration version %d is dirty; resolve it before retrying", version)
	}
	return version, nil
}

// Close releases the migrator and its database handle.
func (m *Migrator) Close() error {
	sourceErr, databaseErr := m.engine.Close()
	return errors.Join(sourceErr, databaseErr, m.db.Close())
}
