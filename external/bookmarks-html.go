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
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"io"
	"strconv"
	"strings"
	"time"
	"tryffel.net/pkg/bookmarker/storage/models"
)

// Content inside these directories will be skipped
// Note: this doesn't take into account the level of folder, any folder matching this will be skipped
var skipFolderNames = []string{
	"Recently Bookmarked",
}

// Folders names that match these are replaced with ""
// Note: this doesn't take into account the level of folder, any folder name matching this will be removed
var removeFolderNames = []string{
	"Bookmarks Toolbar",
}

// Replace dots to avoid unwanted recursion. Change folder names e.g. mypage.com -> mypage-com.
// To disable, set to "."
const replaceDots = "-"

const separator = "."

//ImportBookmarksHtml parses bookmark.html export and returns an array of bookmarks and possible error
func ImportBookmarksHtml(reader io.Reader, mapFoldersProjects bool) ([]*models.Bookmark, error) {
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %v", err)
	}

	// return whether folder complete
	var parseFolder func(node *html.Node) bool

	folder := make([]string, 0, 10)
	folderCounts := &map[string]int{}

	bookmarks := []*models.Bookmark{}

	// Parse single node and its children
	parseFolder = func(node *html.Node) bool {
		foundDir := false
		f := ""
		// Parse folder name
		if node.Type == html.ElementNode && node.Data == "h3" && node.FirstChild != nil {
			f = node.FirstChild.Data

			skip := false
			for _, name := range removeFolderNames {
				if f == name {
					skip = true
					break
				}
			}

			if !skip {
				if strings.Contains(f, separator) {
					f = strings.Replace(f, separator, replaceDots, -1)
				}
				folder = append(folder, f)
				foundDir = true
			}
		}

		// Parse bookmark
		if node.Type == html.TextNode && node.Parent.Type == html.ElementNode && node.Parent.Data == "a" {
			keys := map[string]string{}
			for _, v := range node.Parent.Attr {
				keys[v.Key] = v.Val
			}
			if len(keys) > 0 {
				b := keysTobookmark(keys)
				b.Name = node.Data
				if mapFoldersProjects {
					b.Project = strings.Join(folder, separator)
				}
				bookmarks = append(bookmarks, b)
				(*folderCounts)[strings.Join(folder, ".")] += 1
			}
		}

		// Should we skip current folder
		skipFolder := false
		for _, v := range skipFolderNames {
			if f == v {
				logrus.Info("Skipping folder ", strings.Join(folder, separator))
				skipFolder = true
				break
			}
		}

		// Go through children
		if !skipFolder {
			subDir := false
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				dir := parseFolder(child)
				if dir {
					subDir = true
				}
			}
			if subDir {
				folder = folder[:len(folder)-1]
			}
		}
		return foundDir
	}

	// Parse
	parseFolder(doc)

	// Get folder structure and counts
	dirs := make([]string, len(*folderCounts))
	dirCounts := make([]int, len(*folderCounts))

	i := 0
	for key, value := range *folderCounts {
		dirs[i] = key
		dirCounts[i] = value
		i += 1
	}
	parsed := models.ParseTrees(dirs, dirCounts)
	logrus.Infof("Imported %d bookmarks", len(bookmarks))
	var printChildCount func(project *models.Project)
	printChildCount = func(project *models.Project) {
		logrus.Infof("%s - %d", project.FullName(), project.TotalCount())
		for _, v := range project.Children {
			printChildCount(v)
		}
	}

	for _, v := range parsed {
		printChildCount(v)
	}
	return bookmarks, nil
}

func keysTobookmark(keys map[string]string) *models.Bookmark {
	b := &models.Bookmark{
		Content:   keys["href"],
		CreatedAt: parseUnixTs(keys["add_date"]),
		UpdatedAt: parseUnixTs(keys["last_modified"]),
	}

	tags := keys["tags"]
	if tags != "" {
		b.Tags = strings.Split(tags, ",")
	}

	return b
}

func parseUnixTs(ts string) time.Time {
	if ts == "" {
		return time.Now()
	}

	num, err := strconv.Atoi(ts)
	if err != nil {
		return time.Now()
	}

	return time.Unix(int64(num), 0)
}
