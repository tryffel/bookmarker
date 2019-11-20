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

package modals

import (
	"fmt"
	"github.com/rivo/tview"
	"strings"
	"time"
	"tryffel.net/pkg/bookmarker/config"
)

type ImportData struct {
	File               string
	Tags               []string
	MapFoldersProjects bool
}

type ImportForm struct {
	*tview.Form
	importOngoing bool

	importFunc func(data *ImportData)
	closeFunc  func()
}

func (i *ImportForm) SetDoneFunc(doneFunc func()) {
	i.closeFunc = doneFunc
}

func (i *ImportForm) SetVisible(visible bool) {
}

func (i *ImportForm) SetCreateFunc(importFunc func(data *ImportData)) {
	i.importFunc = importFunc
}

func NewImportForm() *ImportForm {
	i := &ImportForm{Form: tview.NewForm()}

	colors := config.Configuration.Colors.BookmarkForm
	i.SetTitle("Import bookmarks")
	i.SetTitleColor(colors.Text)

	i.SetBorder(true)
	i.SetBorderColor(config.Configuration.Colors.Border)
	i.SetBackgroundColor(colors.Background)
	i.SetLabelColor(colors.Label)
	i.SetFieldBackgroundColor(colors.TextBackground)
	i.SetFieldTextColor(colors.Text)

	i.initForm()
	return i
}

func (i *ImportForm) initForm() {
	now := time.Now()
	ts := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	i.AddInputField("File", "bookmarks.html", 0, nil, nil)
	i.AddInputField("Add tags", fmt.Sprintf("import-%s", ts), 0, nil, nil)
	i.AddCheckbox("Map folders to projects", true, nil)
	i.AddButton("Import", i.doImport)
}

func (i *ImportForm) doImport() {
	if i.importFunc != nil && !i.importOngoing {
		i.importOngoing = true
		data := &ImportData{
			File:               i.GetFormItemByLabel("File").(*tview.InputField).GetText(),
			MapFoldersProjects: i.GetFormItemByLabel("Map folders to projects").(*tview.Checkbox).IsChecked(),
		}
		tags := i.GetFormItemByLabel("Add tags").(*tview.InputField).GetText()
		if tags != "" {
			data.Tags = strings.Split(tags, ",")
		}
		i.importFunc(data)
	}
}

func (i *ImportForm) ImportDone(count int, msg string, ok bool) {
	i.Clear(true)
	if ok {
		i.AddInputField("Status", "Successful", 0, i.denyInput, nil)
		i.AddInputField("Imported", fmt.Sprintf("%d bookmarks", count), 0, i.denyInput, nil)
	} else {
		i.AddInputField("Status", "Failed", 0, i.denyInput, nil)
		i.AddInputField("Imported", fmt.Sprintf("%d bookmarks", count), 0, i.denyInput, nil)
	}
	if msg != "" {
		i.AddInputField("Message", msg, 0, i.denyInput, nil)
	}

	i.AddButton("Close", i.close)
}

func (i *ImportForm) denyInput(string, rune) bool {
	return false
}

//Reset resets form, which must be called before creating new import
func (i *ImportForm) Reset() {
	i.importOngoing = false
	i.Clear(true)
	i.initForm()
}

func (i *ImportForm) close() {
	if i.closeFunc != nil {
		i.closeFunc()
	}
}
