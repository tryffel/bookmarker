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

package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/pkg/bookmarker/config"
	"tryffel.net/pkg/bookmarker/storage/models"
)

const (
	metadataName        = "Name"
	metadataDescription = "Description"
	metadataLink        = "Link"
	metadataProject     = "Project"
	metadataTags        = "Tags"
	metadataCreatedAt   = "Created at"
	metadataUpdatedAt   = "Updated at"
)

var metadataDefaults = []string{metadataName, metadataDescription, metadataLink, metadataProject, metadataTags,
	metadataCreatedAt, metadataUpdatedAt}

//Metadata provides a form-like view to bookmark metadata
type Metadata struct {
	form *tview.Form
	//bookmark
	bookmark *models.Bookmark
	//tmpBookmark for editing commit/rollback
	tmpBookmark *models.Bookmark
	//is edit enabled
	enableEdit bool

	editBtn *tview.Button

	//defaultFields that every bookmark has
	defaultFields      map[string]*tview.InputField
	defaultFieldsArray []*tview.InputField

	doneFunc func(save bool, bookmark *models.Bookmark)
}

func (m *Metadata) Draw(screen tcell.Screen) {
	m.form.Draw(screen)
}

func (m *Metadata) GetRect() (int, int, int, int) {
	return m.form.GetRect()
}

func (m *Metadata) SetRect(x, y, width, height int) {
	m.form.SetRect(x, y, width, height)
}

func (m *Metadata) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()

		if key == tcell.KeyEscape {
			m.doneFunc(false, nil)
		}

		m.form.InputHandler()(event, setFocus)
	}
}

func (m *Metadata) Focus(delegate func(p tview.Primitive)) {
	m.form.Focus(delegate)
}

func (m *Metadata) Blur() {
	m.form.Blur()
}

func (m *Metadata) GetFocusable() tview.Focusable {
	return m.form.GetFocusable()
}

func NewMetadata(doneFunc func(save bool, bookmark *models.Bookmark)) *Metadata {
	m := &Metadata{
		form:          tview.NewForm(),
		bookmark:      nil,
		tmpBookmark:   nil,
		enableEdit:    false,
		editBtn:       tview.NewButton("Edit"),
		doneFunc:      doneFunc,
		defaultFields: map[string]*tview.InputField{},
	}

	colors := config.Configuration.Colors.Metadata

	m.form.SetTitle("Metadata")
	m.form.SetTitleColor(colors.Text)

	m.form.SetBorder(true)
	m.form.SetBorderColor(config.Configuration.Colors.Border)
	m.form.SetBackgroundColor(colors.Background)
	m.form.SetLabelColor(colors.Label)
	m.form.SetFieldBackgroundColor(colors.TextBackground)
	m.form.SetFieldTextColor(colors.Text)

	m.editBtn.SetSelectedFunc(m.toggleEdit)

	m.defaultFieldsArray = make([]*tview.InputField, len(metadataDefaults))
	for i := 0; i < len(metadataDefaults); i++ {
		key := metadataDefaults[i]
		m.defaultFields[key] = tview.NewInputField().SetLabel(key)
		m.defaultFieldsArray[i] = m.defaultFields[key]
	}

	width := 40
	m.defaultFields[metadataName].SetFieldWidth(width)
	m.defaultFields[metadataTags].SetFieldWidth(width)
	m.defaultFields[metadataProject].SetFieldWidth(width)
	m.defaultFields[metadataTags].SetFieldWidth(width)
	m.defaultFields[metadataCreatedAt].SetFieldWidth(width)
	m.defaultFields[metadataUpdatedAt].SetFieldWidth(width)
	return m
}

func (m *Metadata) setData(bookmark *models.Bookmark) {
	m.bookmark = bookmark
	m.form.Clear(true)
	m.initDefaults()
	m.setFields(m.bookmark)
	m.initButtons()
}

func (m *Metadata) setFields(bookmark *models.Bookmark) {
	m.defaultFields[metadataName].SetText(bookmark.Name)
	m.defaultFields[metadataDescription].SetText(bookmark.Description)
	m.defaultFields[metadataLink].SetText(bookmark.Content)
	m.defaultFields[metadataProject].SetText(bookmark.Project)
	m.defaultFields[metadataTags].SetText(bookmark.TagsString(true))
	m.defaultFields[metadataCreatedAt].SetText(bookmark.CreatedAt.Format("2006-01-02 15:04"))
	m.defaultFields[metadataUpdatedAt].SetText(bookmark.UpdatedAt.Format("2006-01-02 15:04"))
}

func (m *Metadata) toggleEdit() {
	if m.enableEdit {
		m.enableEdit = false
		m.editBtn.SetLabel("Edit")
	} else {
		m.enableEdit = true
		m.editBtn.SetLabel("Cancel")
	}
}

func (m *Metadata) initDefaults() {
	for _, field := range m.defaultFieldsArray {
		m.form.AddFormItem(field)
	}
}

func (m *Metadata) initButtons() {
	m.form.AddButton("Edit", m.toggleEdit)
}
