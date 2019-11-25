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

package external

import (
	"golang.org/x/net/html"
	"net/http"
)

type PageMetadata struct {
	Title string
}

func GetPageMetadata(url string) (*PageMetadata, error) {
	metadata := &PageMetadata{}

	resp, err := http.Get(url)
	if err != nil {
		return metadata, err
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return metadata, err
	}

	var iterate func(node *html.Node)

	it := func(node *html.Node) {
		if metadata.Title != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "title" && metadata.Title == "" {
			metadata.Title = node.FirstChild.Data
			return
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			iterate(c)
			if metadata.Title != "" {
				return
			}
		}
	}
	iterate = it
	it(doc)

	return metadata, nil
}
