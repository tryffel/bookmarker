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
		f.Query == ""
}

func (f *Filter) parseTokens(tokens *map[string]StringFilter) error {
	if (*tokens)["query"].Name != "" {
		f.isPlain = true
		f.Query = (*tokens)["query"].Name
		return nil
	}

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
		default:
			f.CustomTags[key] = value
		}
	}
	return nil
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
