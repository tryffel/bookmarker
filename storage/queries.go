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
		var tags string
		err = rows.Scan(&b.Name, &b.Description, &b.Content, &b.Project, &b.CreatedAt, &b.UpdatedAt, &tags)
		if err != nil {
			logrus.Errorf("Scan rows failed: %v", err)
		}
		if len(tags) > 0 {
			t := strings.Split(tags, ",")
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
		var tag string
		var count int
		err = rows.Scan(&tag, &count)
		if err != nil {
			logrus.Error("Error scanning row: %v", err)
		}
		(*results)[tag] = count
	}
	return results, nil
}

func (d *Database) NewBookmark(b *models.Bookmark) error {
	query := `
INSERT INTO 
bookmarks (name, lower_name, description, content, project, created_at, updated_at) 
VALUES (?,?,?,?,?,?,?)`

	_, err := d.conn.Exec(query, b.Name, b.LowerName, b.Description, b.Content, b.Project, b.CreatedAt, b.UpdatedAt)
	return err

}
