package goose

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

// Create writes a new blank migration file.
func CreateWithTemplate(db *sql.DB, dir string, migrationTemplate *template.Template, name, migrationType string) error {
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}

	// Initial version.
	version := nextMigrationNumber(0)

	if last, err := migrations.Last(); err == nil {
		version = nextMigrationNumber(last.Version + 1)
	}

	filename := fmt.Sprintf("%v_%v.%v", version, name, migrationType)

	fpath := filepath.Join(dir, filename)

	tmpl := sqlMigrationTemplate
	if migrationType == "go" {
		tmpl = goSQLMigrationTemplate
	}

	if migrationTemplate != nil {
		tmpl = migrationTemplate
	}

	path, err := writeTemplateToFile(fpath, tmpl, version)
	if err != nil {
		return err
	}

	log.Printf("Created new file: %s\n", path)
	return nil
}

// Create writes a new blank migration file.
func Create(db *sql.DB, dir, name, migrationType string) error {
	return CreateWithTemplate(db, dir, nil, name, migrationType)
}

// Determines the version number of the next migration.
// Source: https://github.com/rails/rails/blob/a9e5457d8cdd1a67a7c6f34a433a9e18057b4222/activerecord/lib/active_record/migration.rb#L910
func nextMigrationNumber(number int64) int64 {
	currentTimeString := time.Now().UTC().Format("20060102150405")

	currentTime, err := strconv.ParseInt(currentTimeString, 10, 64)
	if err != nil || currentTime < number {
		return number
	}

	return currentTime
}

func writeTemplateToFile(path string, t *template.Template, version int64) (string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to create file: %v already exists", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = t.Execute(f, version)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

var sqlMigrationTemplate = template.Must(template.New("goose.sql-migration").Parse(`-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
`))

var goSQLMigrationTemplate = template.Must(template.New("goose.go-migration").Parse(`package migration

import (
	"database/sql"
	"github.com/imjching/goose"
)

func init() {
	goose.AddMigration(Up{{.}}, Down{{.}})
}

func Up{{.}}(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func Down{{.}}(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))
