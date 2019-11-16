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

func TestProject_String(t *testing.T) {
	projects := ParseTrees([]string{"project.test.a"}, []int{5})

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

//This also tests ParseTree
func TestProject_PrintChildren(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		counts []int
		want   string
	}{
		{
			input:  []string{"project.a.b", "project.a.c", "project.b.c"},
			counts: []int{1, 1, 1},
			want:   "project\n   a\n      b\n      c\n   b\n      c",
		},
		{
			input:  []string{"project.a.b", "project.a.b.c.d"},
			counts: []int{1, 1, 1},
			want: `project
   a
      b
         c
            d`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParseTrees(tt.input, tt.counts)[0]
			if got := p.PrintChildren(); got != tt.want {
				t.Errorf("PrintChildren() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_TotalCount(t *testing.T) {
	tests := []struct {
		name    string
		strings []string
		counts  []int
		want    int
	}{
		{
			name:    "1-node",
			strings: []string{"project"},
			counts:  []int{10},
			want:    10,
		},
		{
			name:    "3-nodes",
			strings: []string{"project.a.b", "project.a.c", "project.a.a"},
			counts:  []int{10, 10, 10},
			want:    30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := ParseTrees(tt.strings, tt.counts)[0]

			if got := project.TotalCount(); got != tt.want {
				t.Errorf("TotalCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
