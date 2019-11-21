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
	"reflect"
	"testing"
)

func Test_tokenize(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    *map[string]string
		wantErr bool
	}{{
		name:  "simple query",
		query: "want:one test:two",
		want: &map[string]string{
			"want": "one",
			"test": "two",
		},
		wantErr: false,
	},
		{
			name:    "empty query",
			query:   "",
			want:    &map[string]string{},
			wantErr: false,
		},
		{
			name:  "single value",
			query: "test:one",
			want: &map[string]string{
				"test": "one",
			},
			wantErr: false,
		},
		{
			name:  "value and query",
			query: "test:one two",
			want: &map[string]string{
				"test":  "one",
				"query": "two",
			},
			wantErr: false,
		},
		{
			name:    "key with no value",
			query:   "test: ",
			want:    &map[string]string{},
			wantErr: true,
		},

		/*
				{
				name: "value with space",
				query: "test:'one two'",
				want: &map[string]string{
					"test": "one",
					"query": "two",
				},
				wantErr: false,
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizeQuery(tt.query)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenize() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() return err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    *Filter
		wantErr bool
	}{
		{
			name:  "simple query",
			query: "name:a description:b",
			want: &Filter{
				Name:        StringFilter{Name: "a"},
				Description: StringFilter{Name: "b"},
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFilter(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFilter() got = %v, want %v", got, tt.want)
			}
		})
	}
}
