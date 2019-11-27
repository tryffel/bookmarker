/*
 *   Copyright 2019 Tero Vierimaa
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package storage

import (
	"fmt"
	"strings"
	"time"
)

type StringFilter struct {
	Name   string
	Strict bool
}

//Filter is a filter that represents user defined filterin and sorting
type Filter struct {
	Name          StringFilter
	Description   StringFilter
	Project       StringFilter
	Tags          StringFilter
	Content       StringFilter
	CreatedAfter  time.Time
	CreatedBefore time.Time
	Archived      StringFilter
	CustomTags    map[string]StringFilter
	SortField     string
	SortDir       string
	Query         string
	isPlain       bool
}

//NewFilter parses and constructs new filter based on raw query.
//Query example: "name:asdf* author:'jack' my bookmark"
//Rules: Each parameter is separated by ' ',
//Strict match: enclose word with '\'',
//
func NewFilter(query string) (*Filter, error) {
	f := &Filter{
		CustomTags: map[string]StringFilter{},
	}
	tokens, err := tokenizeQuery(query)
	if err != nil {
		return f, err
	}
	err = f.parseTokens(tokens)
	if err != nil {
		return f, err
	}
	return f, nil
}

func (f *Filter) IsPlainQuery() bool {
	return f.isPlain
}

func (f *Filter) CustomOnly() bool {
	return f.Name.Name == "" &&
		f.Description.Name == "" &&
		f.Project.Name == "" &&
		f.Content.Name == "" &&
		f.Tags.Name == "" &&
		f.Query == "" &&
		f.Archived.Strict == false
}

func (f *Filter) parseTokens(tokens *map[string]StringFilter) error {
	if (*tokens)["query"].Name != "" {
		f.isPlain = true
		f.Query = (*tokens)["query"].Name
		return nil
	}

	var err error

	for key, value := range *tokens {
		switch strings.ToLower(key) {
		case "name":
			f.Name = value
		case "description":
			f.Description = value
		case "project":
			f.Project = value
		case "tags":
			f.Tags = value
			f.Tags.Strict = false
		case "link":
			f.Content = value
		case "after":
			//f.CreatedAfter = value
		case "before":
			//f.CreatedBefore = value
		case "sort":
			f.SortField = value.Name
		case "archived":
			if strings.ToLower(value.Name) == "true" {
				f.Archived.Name = "true"
				f.Archived.Strict = true
			} else if strings.ToLower(value.Name) == "false" {
				f.Archived.Name = "false"
				f.Archived.Strict = true
			} else {
				err = fmt.Errorf("invalid archived format: %v", value.Name)
			}
		default:
			f.CustomTags[key] = value
		}
	}
	return err
}

//Clear clears filter
func (f *Filter) Clear() {
	*f = Filter{}
}

func (f *Filter) IsEmpty() bool {
	return f.CustomOnly() && len(f.CustomTags) == 0
}

//Tokenize tokenizes sentence into key-value pairs
func tokenizeQuery(query string) (*map[string]StringFilter, error) {
	splitChar := " "
	//exactChar := '\''
	keyValueChar := ":"
	queryName := "query"
	result := &map[string]StringFilter{}

	//left := query
	//index := 0

	tokens := strings.Split(query, splitChar)
	if len(tokens) == 0 {
		return result, nil
	} else if len(tokens) == 1 && tokens[0] == "" {
		return result, nil
	}

	for _, token := range tokens {
		t := strings.Split(token, keyValueChar)
		if len(t) == 1 {
			if (*result)[queryName].Name != "" {
				return result, fmt.Errorf("invalid query: %s", token)
			} else {
				(*result)[queryName] = StringFilter{t[0], false}
			}
		} else if len(t) > 2 || t[1] == "" {
			return result, fmt.Errorf("invalid query: '%s'", token)
		} else {
			//runes := []rune(t[1])
			//if runes[0] == exactChar && runes[len(runes)-1] == exactChar {
			(*result)[t[0]] = StringFilter{t[1], false}
		}
	}
	return result, nil
}

//Construct bookmarks query from filter. Return values: query, parameters, error
func (f *Filter) bookmarksQuery() (string, *[]interface{}, error) {
	params := &[]interface{}{}
	queryMetadata := `
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
	b.archived AS archived,
	-- skip tags for now
	'' as tags
FROM bookmarks b
LEFT outer JOIN metadata m on b.id = m.bookmark
WHERE `

	queryBookmark := `
SELECT
	b.id as id,
	b.name AS name,
	b.description AS description,
	b.content as content,
	b.project AS project,
	b.created_at AS created_at,
	b.updated_at AS updated_at,
	b.archived AS archived,
GROUP_CONCAT(t.name) AS tags
FROM bookmarks b
LEFT JOIN bookmark_tags bt ON b.id = bt.bookmark
LEFT JOIN tags t ON bt.tag = t.id
	WHERE `

	queryEnd := `
GROUP BY b.id
ORDER BY b.name ASC
LIMIT 300;`

	queryEndMetadata := ` ) AS a
GROUP BY a.id
ORDER BY a.name ASC
LIMIT 300;`

	query := ""

	metadata := false
	if len(f.CustomTags) > 0 {
		metadata = true
		query = queryMetadata
	}
	i := 0
	for key, filt := range f.CustomTags {
		if i > 0 {
			query += " AND "
		}

		query += "(m.key_lower = ? AND "
		*params = append(*params, strings.ToLower(key))
		if filt.Strict {
			query += "m.value_lower = ?) "
			*params = append(*params, strings.ToLower(filt.Name))
		} else {
			query += "m.value_lower LIKE ?) "
			*params = append(*params, "%"+strings.ToLower(filt.Name)+"%")
		}
		i += 1
	}
	if !f.CustomOnly() {
		if metadata {
			query += "UNION"
		}

		query += queryBookmark

		data := map[string]StringFilter{
			"b.lower_name":  f.Name,
			"b.description": f.Description,
			"b.content":     f.Content,
			"b.project":     f.Project,
			//"b.tags": 		 f.Tags,
		}

		i = 0
		for key, filt := range data {
			if filt.Name != "" {
				if i > 0 {
					query += " AND "
				}
				query += "(" + key
				if filt.Strict {
					query += " = ?)"
					*params = append(*params, strings.ToLower(filt.Name))
				} else {
					query += " LIKE ?)"
					*params = append(*params, "%"+strings.ToLower(filt.Name)+"%")
				}
				i += 1
			}
		}
		if f.Archived.Strict {
			if i > 0 {
				query += " AND "
			}
			query += "b.archived = " + f.Archived.Name
		}
	}

	if metadata {
		query += queryEndMetadata
	} else if !f.CustomOnly() {
		query += queryEnd
	}
	return query, params, nil
}

func (f *Filter) projectsQuery() (string, *[]interface{}, error) {

	params := &[]interface{}{}

	if f.IsEmpty() {
		query := `SELECT
		project,
			count(*) as count
		FROM bookmarks 
		GROUP BY project 
		ORDER BY project ASC;`
		return query, params, nil
	}

	return "", params, fmt.Errorf("not implemented")
}

//BulkUpdateQuery creates a query that modifies bookmarks with filter to value of modifier
func (f *Filter) bulkUpdateQuery(modifier *Modifier) (string, *[]interface{}, error) {
	params := &[]interface{}{}
	query := `
	UPDATE bookmarks SET `
	i := 0

	data := map[string]StringFilter{
		"project":  modifier.Project,
		"archived": modifier.Archived,
	}
	i = 0
	for key, filt := range data {
		if filt.Name != "" {
			if i > 0 {
				query += " AND "
			}
			query += key + " = ?"
			*params = append(*params, strings.ToLower(filt.Name))
			i += 1
		}
	}

	query += " WHERE "
	data = map[string]StringFilter{
		"project":     f.Project,
		"archived":    f.Archived,
		"name":        f.Name,
		"description": f.Description,
	}
	i = 0

	for key, filt := range data {
		if filt.Name != "" {
			if i > 0 {
				query += " AND "
			}
			query += "(" + key
			if filt.Strict {
				query += " = ?)"
				*params = append(*params, strings.ToLower(filt.Name))
			} else {
				query += " LIKE ?)"
				*params = append(*params, "%"+strings.ToLower(filt.Name)+"%")
			}
			i += 1
		}
	}

	return query, params, nil
}
