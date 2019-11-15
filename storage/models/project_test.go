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

package models

import (
	"reflect"
	"testing"
)

func TestParseProject(t *testing.T) {
	object := &Project{
		Name:     "object",
		Children: nil,
		Parent:   nil,
	}
	a := &Project{
		Name:     "a",
		Children: nil,
		Parent:   object,
	}
	b := &Project{
		Name:     "b",
		Children: nil,
		Parent:   object,
	}
	ac := &Project{
		Name:     "c",
		Children: nil,
		Parent:   a,
	}

	ab := &Project{
		Name:     "b",
		Children: nil,
		Parent:   a,
	}

	bc := &Project{
		Name:     "c",
		Children: nil,
		Parent:   b,
	}

	object.Children = []*Project{a, b}
	a.Children = []*Project{ab, ac}
	b.Children = []*Project{bc}

	tests := []struct {
		name string
		text []string
		want []*Project
	}{{
		text: []string{"object.a.b", "object.b.c", "object.a.c"},
		want: []*Project{object},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTrees(tt.text); !reflect.DeepEqual(&got, &tt.want) {
				t.Errorf("ParseProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_String(t *testing.T) {
	projects := ParseTrees([]string{"project.test.a"})

	tests := []struct {
		name    string
		project *Project
		want    string
	}{{
		project: projects[0],
		want:    "project",
	},
		{
			project: projects[0].Children[0],
			want:    "project.test",
		},
		{
			project: projects[0].Children[0].Children[0],
			want:    "project.test.a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.project.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_sortChildren(t *testing.T) {
	tests := []struct {
		name    string
		project *Project
		want    *Project
	}{
		{
			project: &Project{
				Name: "",
				Children: []*Project{{
					Name:     "test-d",
					Children: []*Project{},
				},
					{
						Name:     "test-b",
						Children: []*Project{},
					}},
				Parent: nil,
			},
			want: &Project{
				Name: "",
				Children: []*Project{{
					Name:     "test-b",
					Children: []*Project{},
				},
					{
						Name:     "test-d",
						Children: []*Project{},
					}},
				Parent: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.project.sortChildren(true)

			if !reflect.DeepEqual(tt.project, tt.want) {
				t.Errorf("Project.SortChildren, not sorted")

			}

		})
	}
}
