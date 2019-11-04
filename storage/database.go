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

package storage

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn *sqlx.DB
}

func NewDatabase(file string) (*Database, error) {
	db := &Database{}

	url := fmt.Sprintf("file:%s?_writable_schema=true", file)
	var err error
	db.conn, err = sqlx.Connect("sqlite3", url)

	if err != nil {
		return db, err
	}

	return db, nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) Engine() *sqlx.DB {
	return d.conn
}
