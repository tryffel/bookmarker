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
	"errors"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"strconv"
)

type Database struct {
	conn  *sqlx.DB
	bleve bleve.Index
}

func NewDatabase(file, bleveFile string) (*Database, error) {
	db := &Database{}

	url := fmt.Sprintf("file:%s?_writable_schema=true", file)
	var err error
	db.conn, err = sqlx.Connect("sqlite3", url)

	if err != nil {
		return db, err
	}

	err = db.initBleve(bleveFile)
	return db, err
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) Engine() *sqlx.DB {
	return d.conn
}

func (d *Database) initBleve(file string) error {
	var err error
	d.bleve, err = bleve.Open(file)
	if err != nil {
		if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
			logrus.Infof("Create new bleve database")
			d.bleve, err = bleve.New(file, d.bleveDocumentMapping())
		}
	}
	return err
}

func (d *Database) IndexFts() error {
	var err error

	bookmarks, err := d.GetAllBookmarks()
	if err != nil {
		return err
	}

	for i, v := range bookmarks {
		logrus.Infof("Index %d", i+1)
		id := strconv.Itoa(v.Id)
		err := d.bleve.Index(id, v)
		if err != nil {
			logrus.Errorf("bleve index: %v", err)
		}
	}

	count, err := d.bleve.DocCount()
	if err != nil {
		return err
	}
	logrus.Infof("Bleve contains %d bookmarks", count)
	return nil
}

func (d *Database) bleveDocumentMapping() *mapping.IndexMappingImpl {
	bookmarkMapping := mapping.NewDocumentMapping()
	bMap := bookmarkMapping
	bMap.DefaultAnalyzer = "en"
	name := bleve.NewTextFieldMapping()
	bMap.AddFieldMappingsAt("name", name)
	desc := bleve.NewTextFieldMapping()
	bMap.AddFieldMappingsAt("description", desc)
	url := bleve.NewTextFieldMapping()
	bMap.AddFieldMappingsAt("content", url)
	project := bleve.NewTextFieldMapping()
	bMap.AddFieldMappingsAt("project", project)
	archived := bleve.NewBooleanFieldMapping()
	bMap.AddFieldMappingsAt("archived", archived)
	ts := bleve.NewDateTimeFieldMapping()
	bMap.AddFieldMappingsAt("created_at", ts)
	updatedTs := bleve.NewDateTimeFieldMapping()
	bMap.AddFieldMappingsAt("updated_at", updatedTs)

	mapping := mapping.NewIndexMapping()
	mapping.DefaultField = "bookmark"
	mapping.TypeMapping["bookmark"] = bMap
	return mapping
}
