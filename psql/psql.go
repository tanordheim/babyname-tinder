package psql

import (
	"strings"

	"github.com/jmoiron/sqlx"

	// Import PostgreSQL driver
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Import file migration driver
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/pkg/errors"
)

// NewRepository creates a new PostgreSQL repository.
func NewRepository(connstr string) *Repository {
	db, err := sqlx.Connect("postgres", connstr)
	if err != nil {
		panic(errors.Wrap(err, "Unable to connect to PostgreSQL"))
	}

	migrateDriver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	migrator, err := migrate.NewWithDatabaseInstance("file://psql/migrations", "postgres", migrateDriver)
	if err != nil {
		panic(errors.Wrap(err, "Unable to prepare schema migrations"))
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		panic(errors.Wrap(err, "Unable to perform schema migrations"))
	}

	return &Repository{
		db: db,
	}
}

func getIDForName(name string) string {
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "-", -1)
	return name
}
