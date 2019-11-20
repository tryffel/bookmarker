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
	"github.com/sirupsen/logrus"
	"strings"
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
		err = d.InsertTags(b.Tags)
		if err != nil {
			logrus.Errorf("Insert tags: %v", err)
		} else {
			err = d.UpdateBookmarkTags(b, b.Tags)
			if err != nil {
				logrus.Errorf("Update bookmark tags: %v", err)
			}
		}
	}
	return err
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
		err = d.InsertTags(b.Tags)
		if err != nil {
			logrus.Errorf("Insert tags: %v", err)
		} else {
			err = d.UpdateBookmarkTags(b, b.Tags)
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

func (d *Database) InsertTags(tags []string) error {
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
	_, err := d.conn.Exec(query, args...)
	return err
}

//AddTagsToBookmark adds tags to bookmark. Tags must exist before calling this function
func (d *Database) UpdateBookmarkTags(bookmark *models.Bookmark, tags []string) error {
	query := `DELETE FROM main.bookmark_tags WHERE bookmark_tags.bookmark = ?; INSERT INTO bookmark_tags (bookmark, tag)
	SELECT
		bookmarks.id AS bookmark,
		tag.id AS tag
		FROM bookmarks
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
	query += ")) AS tag WHERE bookmark = ?;"
	_, err := d.conn.Exec(query, args...)
	return err
}
