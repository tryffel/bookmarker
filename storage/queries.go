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
	"github.com/blevesearch/bleve/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/storage/models"
)

//name is any query result that has field 'Name'
type name struct {
	Name string `db:"name"`
}

type query struct {
	start time.Time
	query string
	name  string
}

//Create new logger instance
func beginQuery(Query string, name string) *query {
	q := &query{
		start: time.Now(),
		query: Query,
		name:  name,
	}
	return q
}

func (q *query) log(err error) {
	took := time.Since(q.start)
	ms := took.Milliseconds()

	if err != nil {
		query := strings.Replace(q.query, "\n", " ", -1)
		logrus.Errorf("Sql '%s': %s: %v", q.name, query, err)
	} else {
		logrus.Debugf("Sql %s in %d ms", q.name, ms)
	}
}

//GetAllBookmarks returns all bookmarks filtered by their name.
// cnofig.HideArchived is obeyd and limit is 500
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
`
	queryEnd := `
	
GROUP BY b.id
ORDER BY b.name ASC
LIMIT 500;
`
	if config.Configuration.HideArchived {
		query += " WHERE archived = false"
	}
	query += queryEnd

	logger := beginQuery(query, "get all bookmarks")
	rows, err := d.conn.Query(query)
	if err != nil {
		logger.log(err)
		return nil, err
	}

	bookmarks := make([]*models.Bookmark, 0)
	for rows.Next() {
		var b models.Bookmark
		var tags sql.NullString
		err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &tags)
		if err != nil {
			logrus.Errorf("Scan rows failed: %v", err)
			err = rows.Close()
			break
		}
		if tags.String != "" {
			t := strings.Split(tags.String, ",")
			b.Tags = t
		}
		bookmarks = append(bookmarks, &b)
	}
	logger.log(nil)
	return bookmarks, nil
}

//GetAllProjects gets all projects
// If name is specified, search for that name
// If strict is set to true, then name must match project name exactly,
// Otherwise use wildcards
// If name == "", get all projects
func (d *Database) GetAllProjects(name string, strict bool) ([]*models.Project, error) {
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

	logger := beginQuery(query, "get projects")

	rows, err := d.conn.Query(query, name)
	if err != nil {
		logger.log(err)
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
			err = rows.Close()
			break
		}
		strings = append(strings, project)
		counts = append(counts, count)
	}

	projects := models.ParseTrees(strings, counts)
	logger.log(err)
	return projects, nil
}

//GetAllTags returns all tags
func (d *Database) GetAllTags() (*map[string]int, error) {
	query := `
SELECT
       name,
       COUNT(*) as count
FROM tags
LEFT JOIN bookmark_tags bt ON tags.id = bt.tag
GROUP BY tags.name
ORDER BY tags.name ASC;
`
	logger := beginQuery(query, "get tags")

	rows, err := d.conn.Query(query)
	if err != nil {
		logger.log(err)
		return nil, nil
	}

	results := &map[string]int{}

	for rows.Next() {
		var tag sql.NullString
		var count int
		err = rows.Scan(&tag, &count)
		if err != nil {
			logrus.Errorf("Error scanning row: %v", err)
			err = rows.Close()
			break
		}

		if tag.String != "" {
			(*results)[tag.String] = count
		}
	}
	logger.log(nil)
	return results, nil
}

//NewBookmark stores new bookmark
func (d *Database) NewBookmark(b *models.Bookmark) error {
	query := `
INSERT INTO 
bookmarks (name, lower_name, description, description_lower, content, project, created_at, updated_at, archived) 
VALUES (?,?,?,?,?,?,?,?,?); SELECT last_insert_rowid() FROM bookmarks`

	logger := beginQuery(query, "new bookmark")
	res, err := d.conn.Exec(query, b.Name, b.LowerName, b.Description, strings.ToLower(b.Description), b.Content,
		strings.ToLower(b.Project), b.CreatedAt, b.UpdatedAt, b.Archived)

	if err != nil {
		logger.log(err)
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.log(err)
		return err
	}
	logger.log(nil)
	b.Id = int(id)
	err = d.upsertMetadata(b)
	if err != nil {
		logrus.Errorf("Insert / update bookmark metadata: %v", err)
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
	//Max variables for sqlite is 999, batch of 100 = 900 vars

	numArgs := 9
	batchSize := 100
	total := len(bookmarks)
	imported := 0

	// Get last id
	id := 0
	rows, err := d.conn.Query("SELECT coalesce(max(id),1) FROM main.bookmarks;")
	if err != nil {
		return fmt.Errorf("failed to query last id: %v", err)
	}

	id += 1

	rows.Next()
	err = rows.Scan(&id)
	rows.Close()

	if err != nil {
		return fmt.Errorf("failed to scan last id: %v", err)
	}

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
		bookmarks (name, lower_name, description, description_lower, content, 
		           project, created_at, updated_at, archived) 
		VALUES `

		argList := "(?,?,?,?,?,?,?,?,?)"
		args := make([]interface{}, len(batch)*numArgs)

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
			args[numArgs*i] = v.Name
			args[numArgs*i+1] = v.LowerName
			args[numArgs*i+2] = v.Description
			args[numArgs*i+3] = strings.ToLower(v.Description)
			args[numArgs*i+4] = v.Content
			args[numArgs*i+5] = v.Project
			args[numArgs*i+6] = v.CreatedAt
			args[numArgs*i+7] = v.UpdatedAt
			args[numArgs*i+8] = v.Archived

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

	/*
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

	*/

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

//GetBookmarkMetadata gets metadata related to bookmark
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

	bookmark.FillDefaultMetadata()

	for rows.Next() {
		var key string
		var value string

		err = rows.Scan(&key, &value)
		if err != nil {
			logrus.Errorf("scan rows: %v", err)
			err = rows.Close()
			break
		}

		bookmark.AddMetadata(key, value)

	}
	return nil
}

//GetBookmark returns single bookmark
func (d *Database) GetBookmark(id int) (*models.Bookmark, error) {
	query := `
SELECT
	b.id AS id,
	b.name AS name,
	b.description AS description,
	b.content AS content,
	b.project AS project,
	b.created_at AS created_at,
	b.updated_at AS updated_at,
	b.archived AS archived,
	GROUP_CONCAT(t.name) AS tags
FROM bookmarks b
LEFT JOIN bookmark_tags bt ON b.id = bt.bookmark
LEFT JOIN tags t ON bt.tag = t.id
WHERE b.id = ?
LIMIT 1`

	b := &models.Bookmark{}
	rows, err := d.conn.Query(query, id)
	if err != nil {
		return b, err
	}

	rows.Next()
	var tags sql.NullString
	err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &b.Archived, &tags)

	if tags.String != "" {
		b.Tags = strings.Split(tags.String, ",")
	}
	return b, err
}

//SearchBookmarks searches both bookmarks table and additional metadata fields
// If full text search is enabled, combine those results as well
func (d *Database) SearchBookmarks(text string, isTermQuery bool) ([]*models.Bookmark, error) {

	getString := func(data interface{}, defaultval string) string {
		isStr, ok := data.(string)
		if ok {
			return isStr
		}
		return defaultval
	}

	getBool := func(data interface{}, defaultval bool) bool {
		isbool, ok := data.(bool)
		if ok {
			return isbool
		}
		return defaultval
	}

	var req *bleve.SearchRequest

	if !isTermQuery {
		query := bleve.NewQueryStringQuery(text)
		req = bleve.NewSearchRequest(query)
	} else {
		q := fmt.Sprintf("name:%s description:%s project:%s link:%s", text, text, text, text)
		query := bleve.NewQueryStringQuery(q)
		req = bleve.NewSearchRequest(query)
	}
	req.Fields = []string{"id", "name", "description", "content", "project", "created_at", "updated_at", "archived",
		"tags", "metadata"}
	results, err := d.bleve.Search(req)

	if err != nil {
		return nil, fmt.Errorf("bleve: %v", err)
	}

	bookmarks := make([]*models.Bookmark, results.Total)

	for i, v := range results.Hits {
		bookmark := &models.Bookmark{}
		id, err := strconv.Atoi(v.ID)
		if err != nil {
			logrus.Warningf("bleve returned non int-id: %v", v.ID)
		} else {
			bookmark.Id = id
		}

		bookmark.Name = getString(v.Fields["name"], "")
		bookmark.Description = getString(v.Fields["description"], "")
		bookmark.Content = getString(v.Fields["content"], "")
		bookmark.Project = getString(v.Fields["project"], "")
		bookmark.Archived = getBool(v.Fields["archived"], false)
		bookmarks[i] = bookmark
	}
	return bookmarks, nil
}

//UpdateBookmark updates all fields on bookmark
func (d *Database) UpdateBookmark(b *models.Bookmark) error {
	query := `
UPDATE bookmarks SEt
		name = ?,
		lower_name = ?,
		description = ?,
        description_lower = ?,
		content = ?,
		project = ?,
		updated_at = ?,
		archived = ?
WHERE id = ?;
`
	_, err := d.conn.Exec(query, b.Name, b.LowerName, b.Description, strings.ToLower(b.Description),
		b.Content, strings.ToLower(b.Project), b.UpdatedAt, b.Archived, b.Id)

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

	logger := beginQuery(query, "update/insert bookmark metadata")
	var err error

	for key, value := range *b.Metadata {
		keyLower := strings.ToLower(key)
		valueLower := strings.ToLower(value)

		_, err = d.conn.Exec(query, b.Id, key, keyLower, value, valueLower, value, valueLower, b.Id)
		if err != nil {
			logrus.Errorf("Failed to insert/update metadata: %v", err)
		}
	}
	logger.log(err)
	return nil
}

//InsertTag inserts tag for bookmark
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

//DeleteBookmark deletes bookmark and its metadata
func (d *Database) DeleteBookmark(bookmark *models.Bookmark) error {
	query := `
DELETE FROM bookmarks
WHERE bookmarks.id = ?`

	_, err := d.conn.Exec(query, bookmark.Id)
	return err
}

//GetStatistics gets various stats related to stored bookmarks
func (d *Database) GetStatistics() (*Statistics, error) {
	s := &Statistics{}

	query := `
	SELECT
		COUNT(b.id) AS bookmarks,
		(SELECT count(id) FROM bookmarks WHERE archived=true) AS archived,
		COUNT(DISTINCT(b.project)) AS projects
	FROM bookmarks b`

	rows, err := d.conn.Query(query)
	if err != nil {
		e := rows.Close()
		if e != nil {
			err = fmt.Errorf("%v; %v", err, e)
		}
		return s, err
	}

	rows.Next()
	err = rows.Scan(&s.Bookmarks, &s.Archived, &s.Projects)
	if err != nil {
		return s, err
	}

	query = `
SELECT b.created_at
FROM bookmarks b
ORDER BY b.created_at DESC
LIMIT 1
`
	rows, err = d.conn.Query(query)
	if err != nil {
		rows.Close()
		return s, err
	}
	rows.Next()
	err = rows.Scan(&s.LastBookmark)

	s.FullTextSearchSupported, err = d.FullTextSearchSupported()

	s.MetadataKeys, err = d.GetMetadataKeys()
	return s, err
}

//FilterBookmarks applies given filter to return matching bookmarks
func (d *Database) FilterBookmarks(filter *Filter) ([]*models.Bookmark, error) {
	query, params, err := filter.bookmarksQuery()
	if err != nil {
		return nil, err
	}

	logger := beginQuery(query, "filter bookmarks")

	rows, err := d.conn.Query(query, *params...)
	if err != nil {
		logger.log(err)
		return nil, err
	}

	bookmarks := []*models.Bookmark{}
	for rows.Next() {
		var tag sql.NullString
		b := models.Bookmark{}

		err := rows.Scan(&b.Id, &b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &b.Archived, &tag)
		if err != nil {
			logrus.Errorf("scan bookmark rows: %v", err)
			err = rows.Close()
			break
		}

		if tag.String != "" {
			b.Tags = strings.Split(tag.String, ",")
		}

		bookmarks = append(bookmarks, &b)
	}
	logger.log(err)
	return bookmarks, nil
}

//SearchKeyValue searches any key-value item for bookmark
func (d *Database) SearchKeyValue(key, value string) ([]string, error) {
	key = strings.ToLower(key)
	value = "%" + strings.ToLower(value) + "%"
	limit := config.Configuration.AutoCompleteMaxResults

	query := `
SELECT value_lower
FROM metadata
WHERE key_lower = ? 
AND value_lower LIKE ?
GROUP BY value_lower 
ORDER BY value_lower ASC
LIMIT ?;`

	//TODO: user filter for queries
	if key == "project" {
		query = `
SELECT project
FROM bookmarks
WHERE project LIKE ? 
GROUP BY project
ORDER BY project ASC
LIMIT ?;`
	}

	logger := beginQuery(query, "search metadata")
	results := make([]string, 0, 0)

	var rows *sql.Rows
	var err error

	if key == "project" {
		rows, err = d.conn.Query(query, value, limit)
	} else {
		rows, err = d.conn.Query(query, key, value, limit)
	}
	if err != nil {
		logger.log(err)
		return results, err
	}

	for rows.Next() {
		var val string
		err = rows.Scan(&val)
		if err != nil {
			logrus.Errorf("Scan value: %v", err)
			err = rows.Close()
			break
		}

		results = append(results, val)
	}
	logger.log(nil)
	return results, nil
}

//Bulk modify modifies multple bookmarks defined with filter to state defined in modifier
func (d *Database) BulkModify(filter *Filter, modifier *Modifier) (int, error) {
	query, params, err := filter.bulkUpdateQuery(modifier)
	if err != nil {
		return 0, err
	}

	res, err := d.conn.Exec(query, *params...)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	return int(count), err
}

// FilterProject filters projects by given filter. If only filter.Project is defined
// filter project by that. Otherwise filter projects by bookmarks that match given filter
func (d *Database) FilterProject(filter *Filter) ([]*models.Project, error) {
	query := `
	SELECT 
	project,
	count(id) as count
	FROM (
	`

	q, params, err := filter.bookmarksQuery()
	if err != nil {
		return nil, err
	}

	query += q
	query += ") GROUP BY project ORDER BY project ASC"

	logger := beginQuery(query, "filter projects")

	rows, err := d.conn.Query(query, *params...)
	if err != nil {
		logger.log(err)
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
			err = rows.Close()
			break
		}
		strings = append(strings, project)
		counts = append(counts, count)
	}

	projects := models.ParseTrees(strings, counts)
	logger.log(err)
	return projects, nil
}

//FullTextSearchSupported returns whether sqlite FTS5-module is enabled
func (d *Database) FullTextSearchSupported() (bool, error) {
	return true, nil
}

//GetMetadataKeys returns all known metadata keys
func (d *Database) GetMetadataKeys() ([]string, error) {
	query := `
SELECT key
FROM metadata
GROUP BY key
ORDER BY key ASC;
`

	log := beginQuery(query, "get all metadata keys")

	results := make([]string, 0)

	rows, err := d.conn.Query(query)
	if err != nil {
		log.log(err)
		return results, err
	}

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			rows.Close()
			break
		}
		results = append(results, key)
	}
	log.log(err)
	return results, err
}
