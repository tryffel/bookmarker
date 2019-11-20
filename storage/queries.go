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
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"tryffel.net/pkg/bookmarker/storage/models"
)

//name is any query result that has field 'Name'
type name struct {
	Name string `db:"name"`
}

func (d *Database) GetAllBookmarks() ([]*models.Bookmark, error) {
	query := `
SELECT
    b.id as id,
    b.name AS name, 
    b.description AS description, 
    b.content as content, 
    b.project AS project, 
    b.created_at AS created_at,
    b.updated_at AS updated_at,
    GROUP_CONCAT(t.name) AS tags
FROM bookmarks b
LEFT JOIN bookmark_tags bt ON b.id = bt.bookmark
LEFT JOIN tags t ON bt.tag = t.id
GROUP BY b.id
ORDER BY b.name ASC
LIMIT 100;
`

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, err
	}

	bookmarks := make([]*models.Bookmark, 0)

	for rows.Next() {
		var b models.Bookmark
		var tags sql.NullString
		err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &tags)
		if err != nil {
			logrus.Errorf("Scan rows failed: %v", err)
		}
		if tags.String != "" {
			t := strings.Split(tags.String, ",")
			b.Tags = t
		}
		bookmarks = append(bookmarks, &b)
	}

	return bookmarks, nil
}

//GetProjects gets all projects
// If name is specified, search for that name
// If strict is set to true, then name must match project name exactly,
// Otherwise use wildcards
// If name == "", get all projects
func (d *Database) GetProjects(name string, strict bool) ([]*models.Project, error) {
	query := `
SELECT 
    project,
	count(*) as count
FROM bookmarks `

	if name != "" {
		query += " WHERE project "
		if strict {
			query += "= ?"
		} else {
			query += "LIKE '%?%' "
		}
	}

	query += " GROUP BY project ORDER BY project ASC;"

	rows, err := d.conn.Query(query, name)
	if err != nil {
		return nil, err
	}

	strings := make([]string, 0)
	counts := make([]int, 0)

	for rows.Next() {
		var project = ""
		var count = 0

		err := rows.Scan(&project, &count)
		if err != nil {
			logrus.Errorf("scan rows: %v", err)
		}
		strings = append(strings, project)
		counts = append(counts, count)
	}

	projects := models.ParseTrees(strings, counts)
	return projects, nil
}

func (d *Database) GetTags() (*map[string]int, error) {
	query := `
SELECT
       name,
       COUNT(*) as count
FROM tags
LEFT JOIN bookmark_tags bt ON tags.id = bt.tag
GROUP BY tags.name
ORDER BY tags.name ASC;
`
	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, nil
	}

	results := &map[string]int{}

	for rows.Next() {
		var tag sql.NullString
		var count int
		err = rows.Scan(&tag, &count)
		if err != nil {
			logrus.Error("Error scanning row: %v", err)
		}

		if tag.String != "" {
			(*results)[tag.String] = count
		}
	}
	return results, nil
}

func (d *Database) NewBookmark(b *models.Bookmark) error {
	query := `
INSERT INTO 
bookmarks (name, lower_name, description, content, project, created_at, updated_at) 
VALUES (?,?,?,?,?,?,?); SELECT last_insert_rowid() FROM bookmarks`

	res, err := d.conn.Exec(query, b.Name, b.LowerName, b.Description, b.Content, strings.ToLower(b.Project),
		b.CreatedAt, b.UpdatedAt)

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		logrus.Error(err)
		return err
	}
	b.Id = int(id)
	err = d.upsertMetadata(b)
	if err != nil {
		logrus.Error("Insert / update bookmark metadata: %v", err)
	}
	if len(b.Tags) > 0 {
		err = d.InsertTags(b.Tags, nil)
		if err != nil {
			logrus.Errorf("Insert tags: %v", err)
		} else {
			err = d.UpdateBookmarkTags(b, b.Tags, nil)
			if err != nil {
				logrus.Errorf("Update bookmark tags: %v", err)
			}
		}
	}
	return err
}

//NewBookmarks creates batch of new bookmarks. Bookmark ids are not collected and need to be queried separately
// AddTags allows defining any custom tags that are assigned to all bookmarks
func (d *Database) NewBookmarks(bookmarks []*models.Bookmark, AddTags []string) error {
	//Max variables for sqlite is 999, batch of 100 = 700 vars
	batchSize := 100
	total := len(bookmarks)
	imported := 0

	// Get last id
	id := 0
	rows, err := d.conn.Query("SELECT max(id) FROM main.bookmarks;")
	if err != nil {
		return fmt.Errorf("failed to query last id: %v", err)
	}

	id += 1

	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to scan last id: %v", err)

	}
	err = rows.Close()

	tx, err := d.conn.Beginx()
	if err != nil {
		return fmt.Errorf("start transaction: %v", err)
	}

	tagsMap := map[string]bool{}

	// Import bookmarks in batches
	for {
		var batch []*models.Bookmark

		if imported+batchSize < total {
			batch = bookmarks[imported : imported+batchSize]
		} else {
			batch = bookmarks[imported:]
		}

		query := `
		INSERT INTO 
		bookmarks (name, lower_name, description, content, project, created_at, updated_at) 
		VALUES `

		argList := "(?,?,?,?,?,?,?)"
		args := make([]interface{}, len(batch)*7)

		// Parse each bookmark, put tags to map, put bookmark to args list
		for i, v := range batch {
			if i > 0 {
				query += ","
			}
			query += argList
			v.Id = id
			id += 1

			if v.LowerName == "" {
				v.LowerName = strings.ToLower(v.Name)
			}
			if v.CreatedAt == time.Unix(0, 0) {
				v.CreatedAt = time.Now()
			}
			if v.UpdatedAt == time.Unix(0, 0) {
				v.UpdatedAt = time.Now()
			}

			if len(AddTags) > 0 {
				v.AddTags(AddTags)
			}
			args[7*i] = v.Name
			args[7*i+1] = v.LowerName
			args[7*i+2] = v.Description
			args[7*i+3] = v.Content
			args[7*i+4] = v.Project
			args[7*i+5] = v.CreatedAt
			args[7*i+6] = v.UpdatedAt

			if len(v.Tags) > 0 {
				for _, t := range v.Tags {
					tagsMap[t] = true
				}
			}
		}
		query += ";"

		res, err := tx.Exec(query, args...)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert bookmarks: %v", err)
		}

		rows, err := res.RowsAffected()
		if int(rows) != len(batch) {
			logrus.Warning("Some bookmarks were not stored properly during import, probably due to duplicate")
		}
		imported += len(batch)
		if imported == total {
			break
		}
	}

	//Tags
	tags := make([]string, len(tagsMap))
	i := 0
	for key, _ := range tagsMap {
		tags[i] = key
		i += 1
	}
	err = d.InsertTags(tags, tx)

	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("insert tags: %v", err)
	}

	//Tags bookmarks relations
	//Add tags one bookmark at a time for now
	for _, v := range bookmarks {
		if len(v.Tags) > 0 {
			err = d.UpdateBookmarkTags(v, v.Tags, tx)
			if err != nil {
				break
			}
		}
	}

	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("udpate bookmark tags relations: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	return nil
}

func (d *Database) GetBookmarkMetadata(bookmark *models.Bookmark) error {
	query := `
SELECT key, value FROM 
metadata WHERE metadata.bookmark = ?
ORDER BY metadata.bookmark ASC, 
         metadata.key_lower ASC;
`

	rows, err := d.conn.Query(query, bookmark.Id)
	if err != nil {
		return err
	}

	bookmark.Metadata = &map[string]string{}
	bookmark.MetadataKeys = &[]string{}

	for rows.Next() {
		var key string
		var value string

		err = rows.Scan(&key, &value)
		if err != nil {
			logrus.Errorf("scan rows: &v", err)
		}

		(*bookmark.Metadata)[key] = value
		*bookmark.MetadataKeys = append(*bookmark.MetadataKeys, key)
	}
	return nil
}

//SearchBookmarks searches both bookmarks table and additional metadata fields
func (d *Database) SearchBookmarks(text string) ([]*models.Bookmark, error) {
	text = "%" + text + "%"

	query := `
-- metadata
SELECT * FROM 
(
	SELECT
    	b.id AS id,
    	b.name AS name,
    	b.description AS description,
    	b.content AS content,
    	b.project AS project,
    	b.created_at AS created_at,
    	b.updated_at AS updated_at,
       	-- skip tags for now
    	'' as tags
	FROM bookmarks b
         LEFT outer JOIN metadata m on b.id = m.bookmark
	WHERE m.value_lower like ?
	UNION
	-- bookmarks with tags
	SELECT
		b.id as id,
		b.name AS name,
		b.description AS description,
		b.content as content,
		b.project AS project,
		b.created_at AS created_at,
		b.updated_at AS updated_at,
		GROUP_CONCAT(t.name) AS tags
	FROM bookmarks b
		LEFT JOIN bookmark_tags bt ON b.id = bt.bookmark
		LEFT JOIN tags t ON bt.tag = t.id
	WHERE b.lower_name LIKE ?
		OR b.description LIKE ?
		OR b.content LIKE ?
		OR b.project LIKE ?
		OR t.name LIKE ?
) AS a
GROUP BY a.id
ORDER BY a.name ASC
LIMIT 300;
`

	rows, err := d.conn.Query(query, text, text, text, text, text, text, text)
	if err != nil {
		return nil, err
	}

	bookmarks := []*models.Bookmark{}

	for rows.Next() {
		var tag sql.NullString
		b := models.Bookmark{}

		err := rows.Scan(&b.Id, &b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &tag)
		if err != nil {
			logrus.Errorf("scan bookmark rows: %v", err)
		}

		if tag.String != "" {
			b.Tags = strings.Split(tag.String, ",")
		}

		bookmarks = append(bookmarks, &b)
	}
	return bookmarks, nil
}

func (d *Database) UpdateBookmark(b *models.Bookmark) error {
	query := `
UPDATE bookmarks SEt
		name = ?,
		lower_name = ?,
		content = ?,
		project = ?,
		updated_at = ?
WHERE id = ?;
`
	_, err := d.conn.Exec(query, b.Name, b.LowerName, b.Content, strings.ToLower(b.Project), b.UpdatedAt, b.Id)

	if err != nil {
		return err
	}

	err = d.upsertMetadata(b)
	if err != nil {
		return err
	}

	if len(b.Tags) > 0 {
		err = d.InsertTags(b.Tags, nil)
		if err != nil {
			logrus.Errorf("Insert tags: %v", err)
		} else {
			err = d.UpdateBookmarkTags(b, b.Tags, nil)
			if err != nil {
				logrus.Errorf("Update bookmark tags: %v", err)
			}
		}
	}

	return err
}

//UpsertMetadata upserts (insert / update) metadata
func (d *Database) upsertMetadata(b *models.Bookmark) error {
	query := `
INSERT INTO metadata
(bookmark, key, key_lower, value, value_lower)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(bookmark, key_lower) DO UPDATE SET
value = ?, value_lower = ?
WHERE bookmark = ?
`

	for key, value := range *b.Metadata {
		keyLower := strings.ToLower(key)
		valueLower := strings.ToLower(value)

		_, err := d.conn.Exec(query, b.Id, key, keyLower, value, valueLower, value, valueLower, b.Id)
		if err != nil {
			logrus.Errorf("Failed to insert/update metadata: %v", err)
		}
	}
	return nil
}

//GetProjectBookmarks gets all bookmarks with given project.
// If strict = true, project must match exatly, else all children are returned also
func (d *Database) GetProjectBookmarks(project string, strict bool) ([]*models.Bookmark, error) {

	query := `
SELECT
    b.id as id,
    b.name AS name, 
    b.description AS description, 
    b.content as content, 
    b.project AS project, 
    b.created_at AS created_at,
    b.updated_at AS updated_at,
    GROUP_CONCAT(t.name) AS tags
FROM bookmarks b
LEFT JOIN bookmark_tags bt ON b.id = bt.bookmark
LEFT JOIN tags t ON bt.tag = t.id 
WHERE b.project `

	if !strict {
		query += "LIKE ? "
		project = "%" + project + "%"
	} else {
		query += "= ? "
	}
	query += `
GROUP BY b.id
ORDER BY b.name ASC
LIMIT 100;
`
	rows, err := d.conn.Query(query, project)
	if err != nil {
		return nil, err
	}

	bookmarks := make([]*models.Bookmark, 0)
	for rows.Next() {
		var b models.Bookmark
		var tags sql.NullString
		err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &tags)
		if err != nil {
			logrus.Errorf("Scan rows failed: %v", err)
		}
		if tags.String != "" {
			t := strings.Split(tags.String, ",")
			b.Tags = t
		}
		bookmarks = append(bookmarks, &b)
	}
	return bookmarks, nil
}

func (d *Database) InsertTags(tags []string, tx *sqlx.Tx) error {
	query := "INSERT INTO tags (name) VALUES "

	args := make([]interface{}, len(tags))
	for i, v := range tags {
		if i > 0 {
			query += ","
		}
		query += "(?) "
		args[i] = v
	}
	query += " ON CONFLICT (name) DO NOTHING;"

	var err error

	if tx != nil {
		_, err = tx.Exec(query, args...)
	} else {
		_, err = d.conn.Exec(query, args...)
	}
	return err
}

//AddTagsToBookmark adds tags to bookmark. Tags must exist before calling this function.
//If tx is nil, use database connection directly
func (d *Database) UpdateBookmarkTags(bookmark *models.Bookmark, tags []string, tx *sqlx.Tx) error {
	query := `DELETE FROM main.bookmark_tags WHERE bookmark_tags.bookmark = ?; INSERT INTO bookmark_tags (bookmark, tag)
	SELECT
		bookmark,
		tag.id AS tag
		FROM (SELECT ? AS bookmark) 
		LEFT JOIN (SELECT id FROM tags WHERE name IN (`

	args := make([]interface{}, len(tags)+2)
	args[0] = bookmark.Id
	for i, v := range tags {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i+1] = v
	}

	args[len(args)-1] = bookmark.Id
	query += ")) AS tag;"

	var err error
	if tx != nil {
		_, err = tx.Exec(query, args...)
	} else {
		_, err = d.conn.Exec(query, args...)
	}
	return err
}

//RenameProject renames project and all its children.
// e.g. old: my-awesome-project, new: project:
// results in my-awesome-project.a -> project.a
func (d *Database) RenameProject(old string, new string) error {
	return nil

}
