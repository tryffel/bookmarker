/*
 *
 *  Copyright 2019 Tero Vierimaa
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package migrations

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

var BookmarkerMigrations = []Migrator{
	&Migration{
		Name:   "initial schema",
		Level:  1,
		Schema: v1,
	},
}

type Schema struct {
	Level     int       `db:"level"`
	Success   int       `db:"success"`
	Timestamp time.Time `db:"timestamp"`
	TookMs    int       `db:"took_ms"`
}

// Migrator describes single migration level
type Migrator interface {
	// Get migration name
	MName() string
	// Get migration level
	MLevel() int
	// Get valid sql string to execute
	MSchema() string
}

// Migration implements migrator
type Migration struct {
	Name   string
	Level  int
	Schema string
}

func (m *Migration) MName() string {
	return m.Name
}

func (m *Migration) MLevel() int {
	return m.Level
}

func (m *Migration) MSchema() string {
	return m.Schema
}

// Migrate runs given migrations
func Migrate(db *sqlx.DB, migrations []Migrator) error {
	current, err := CurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %v", err)
	}

	if current.Level == 0 {
		_, err := db.Exec(`
CREATE TABLE "schemas" (
	"level"	INTEGER,
	"success"	INTEGER NOT NULL,
	"timestamp"	TIMESTAMP NOT NULL,
	"took_ms"	INTEGER NOT NULL,
	PRIMARY KEY("level")
);
`)

		if err != nil {
			return fmt.Errorf("failed create schema table: %v", err)
		}

	} else {
		if current.Success == 0 {
			return fmt.Errorf("previous migration has failed")
		}
	}

	if current.Level == migrations[len(migrations)-1].MLevel() {
		logrus.Debug("No new migrations to run")
		return nil
	}

	lastLevel := current.Level
	for _, v := range migrations[current.Level:] {
		logrus.Warningf("Migrating database schema %d -> %d", lastLevel, v.MLevel())
		err := migrateSingle(db, v)
		if err != nil {
			return fmt.Errorf("failed to run migrations: %v", err)
		}
		lastLevel = v.MLevel()
	}
	return nil
}

// Run single migration
func migrateSingle(db *sqlx.DB, migration Migrator) error {
	start := time.Now()
	_, merr := db.Exec(migration.MSchema())

	s := &Schema{
		Level:     migration.MLevel(),
		Timestamp: time.Now(),
		TookMs:    int(time.Since(start).Nanoseconds() / 1000000),
	}

	if merr == nil {
		s.Success = 1
	} else {
		s.Success = 0
	}

	_, err := db.Exec("INSERT INTO schemas (level, success, timestamp, took_ms) "+
		"VALUES ($1, $2, $3, $4)", s.Level, s.Success, s.Timestamp, s.TookMs)

	if err != nil {
		return fmt.Errorf("migration failed: insert schema: %v", merr)
	}
	return nil
}

// CurrentVersion returns current version
func CurrentVersion(db *sqlx.DB) (Schema, error) {
	current := Schema{}
	err := db.Get(&current, "SELECT * FROM schemas ORDER BY level DESC LIMIT 1")

	if err != nil {
		e := err.Error()

		if err.Error() == "relation \"schemas\" does not exist" || e == "no such table: schemas" {
			return Schema{
				Level:     0,
				Success:   0,
				Timestamp: time.Time{},
				TookMs:    0,
			}, nil
		}

		return Schema{}, fmt.Errorf("failed to query schema: %v", err)

	}
	return current, nil
}
