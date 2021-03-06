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
	"strings"
)

type Modifier struct {
	Project    StringFilter
	Tags       StringFilter
	Archived   StringFilter
	CustomTags map[string]StringFilter
}

func NewModifier(key, value string) (*Modifier, error) {
	m := &Modifier{}

	switch strings.ToLower(key) {
	case "project":
		m.Project.Name = value
		m.Project.Strict = true
	case "tags":
	case "archived":
		m.Archived.Name = value
	default:
	}

	return m, nil
}
