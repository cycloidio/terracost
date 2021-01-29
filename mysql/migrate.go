package mysql

import (
	"context"
	"database/sql"

	"github.com/lopezator/migrator"

	"github.com/cycloidio/terracost/mysql/migrations"
)

// Migrate runs the migrations on the provided DB using the provided table to track them.
func Migrate(ctx context.Context, db *sql.DB, table string) error {
	ms := make([]interface{}, 0, len(migrations.Migrations))
	for _, m := range migrations.Migrations {
		m := m
		ms = append(ms, &migrator.Migration{
			Name: m.Name,
			Func: func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
					return err
				}
				return nil
			},
		})
	}

	mig, err := migrator.New(migrator.TableName(table), migrator.Migrations(ms...))
	if err != nil {
		return err
	}

	if err := mig.Migrate(db); err != nil {
		return err
	}

	return nil
}
