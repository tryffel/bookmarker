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
	"testing"
)

func TestBookmark_ContentDomain(t *testing.T) {
	type fields struct {
		Content string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "valid url",
			fields: fields{Content: "https://test.net/some/blog?a=1"},
			want:   "test.net",
		},
		{
			name:   "invalid url",
			fields: fields{Content: "https://tes t.net/some/blog?a=1"},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmark{
				Content: tt.fields.Content,
			}
			if got := b.ContentDomain(); got != tt.want {
				t.Errorf("ContentDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBookmark_TagsString(t *testing.T) {
	type fields struct {
		Tags []string
	}
	type args struct {
		spaces bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "no tags",
			fields: fields{
				Tags: []string{},
			},
			args: args{spaces: false},
			want: "",
		},
		{
			name: "tags no space",
			fields: fields{
				Tags: []string{"a", "b", "c"},
			},
			args: args{spaces: false},
			want: "a,b,c",
		},
		{
			name: "tags with space",
			fields: fields{
				Tags: []string{"a", "b", "c"},
			},
			args: args{spaces: true},
			want: "a, b, c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmark{
				Tags: tt.fields.Tags,
			}
			if got := b.TagsString(tt.args.spaces); got != tt.want {
				t.Errorf("TagsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
