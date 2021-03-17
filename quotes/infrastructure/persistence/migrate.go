package persistence

import (
	"github.com/go-pg/migrations/v7"
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

func Migrate(db *pg.DB, path string) error {
	migrationCollection := migrations.NewCollection()
	_, _, err := migrationCollection.Run(db, "init")
	if err != nil {
		return errors.Wrap(err, "Migrate failed init")
	}

	err = migrationCollection.DiscoverSQLMigrations(path)
	if err != nil {
		return errors.Wrap(err, "Migrate failed discover")
	}

	_, _, err = migrationCollection.Run(db, "up")
	if err != nil {
		return errors.Wrap(err, "Migrate failed run")
	}

	return nil
}
