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

import "tryffel.net/pkg/bookmarker/storage/models"

type project struct {
	Project string `db:"project"`
}

func (d *Database) GetAllBookmarks() ([]*models.Bookmark, error) {
	query := `
SELECT name, description, content, project, created_at, updated_at
FROM bookmarks
ORDER BY lower_name ASC
LIMIT 100;`

	p := make([]models.Bookmark, 0)
	err := d.conn.Select(&p, query)
	if err != nil {
		return nil, err
	}

	bookmarks := make([]models.Bookmark, 0, len(p))
	ptrs := make([]*models.Bookmark, 0, len(p))

	copy(bookmarks, p)

	for i := 0; i < len(p); i++ {
		bookmarks = append(bookmarks, p[i])
		ptrs = append(ptrs, &p[i])
		//ptrs[i] = &bookmarks[i]
	}

	return ptrs, nil
}

//GetProjects gets all projects
// If name is specified, search for that name
// If strict is set to true, then name must match project name exactly,
// Otherwise use wildcards
// If name == "", get all projects
func (d *Database) GetProjects(name string, strict bool) ([]*models.Project, error) {
	query := `
SELECT DISTINCT(project) as project
FROM bookmarks`

	if name != "" {
		query += " WHERE project "
		if strict {
			query += "= ?"
		} else {
			query += "LIKE '%?%' "
		}
	}

	p := []project{}
	err := d.conn.Select(&p, query, name)
	if err != nil {
		return nil, err
	}

	strings := make([]string, len(p))
	for i, v := range p {
		strings[i] = v.Project
	}

	projects := models.ParseTrees(strings)
	return projects, nil
}

func (d *Database) NewBookmark(b *models.Bookmark) error {
	query := `
INSERT INTO 
bookmarks (name, lower_name, description, content, project, created_at, updated_at) 
VALUES (?,?,?,?,?,?,?)`

	_, err := d.conn.Exec(query, b.Name, b.LowerName, b.Description, b.Content, b.Project, b.CreatedAt, b.UpdatedAt)
	return err

}
