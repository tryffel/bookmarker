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
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/external"
	"tryffel.net/go/bookmarker/storage/models"
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

	//defaultFields that every bookmark has
	defaultFields      map[string]*tview.InputField
	defaultFieldsArray []*tview.InputField

	customFields *map[string]*tview.InputField
	customKeys   *[]string
	archived     *tview.Checkbox

	doneFunc func(save bool, bookmark *models.Bookmark) bool
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

func NewMetadata(doneFunc func(save bool, bookmark *models.Bookmark) bool) *Metadata {
	m := &Metadata{
		form:          tview.NewForm(),
		bookmark:      nil,
		tmpBookmark:   nil,
		enableEdit:    false,
		doneFunc:      doneFunc,
		defaultFields: map[string]*tview.InputField{},
		archived:      tview.NewCheckbox(),
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

	m.defaultFieldsArray = make([]*tview.InputField, len(metadataDefaults))
	for i := 0; i < len(metadataDefaults); i++ {
		key := metadataDefaults[i]
		m.defaultFields[key] = tview.NewInputField().SetLabel(key)
		m.defaultFieldsArray[i] = m.defaultFields[key]
	}

	width := 40
	m.defaultFields[metadataName].SetFieldWidth(width).SetAcceptanceFunc(m.editEnabled)
	m.defaultFields[metadataTags].SetFieldWidth(width).SetAcceptanceFunc(m.editEnabled)
	m.defaultFields[metadataProject].SetFieldWidth(width).SetAcceptanceFunc(m.editEnabled)
	m.defaultFields[metadataTags].SetFieldWidth(width).SetAcceptanceFunc(m.editEnabled)
	m.defaultFields[metadataCreatedAt].SetFieldWidth(width).SetAcceptanceFunc(m.editEnabled)
	m.defaultFields[metadataUpdatedAt].SetFieldWidth(width).SetAcceptanceFunc(m.editEnabled)
	m.archived.SetLabel("Archived")

	m.initDefaults()
	return m
}

func (m *Metadata) setData(bookmark *models.Bookmark) {
	bookmark.FillDefaultMetadata()
	m.bookmark = bookmark
	m.form.Clear(true)
	m.customFields = &map[string]*tview.InputField{}
	m.customKeys = bookmark.MetadataKeys
	m.initDefaults()
	m.setFields(m.bookmark)
	m.initCustomFields()
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
	m.archived.SetChecked(bookmark.Archived)
}

func (m *Metadata) toggleEdit() {
	if m.enableEdit {
		m.form.SetFieldBackgroundColor(config.Configuration.Colors.Metadata.TextBackground)
		m.enableEdit = false
	} else {
		m.enableEdit = true
		m.form.SetFieldBackgroundColor(config.Configuration.Colors.Metadata.BackgroundEditable)
		ind := m.form.GetButtonIndex("Edit")
		m.form.RemoveButton(ind)

		m.form.AddButton("Save", m.save)
		m.form.AddButton("Cancel", m.cancel)
		m.form.AddButton("Get title", m.getTitle)
	}
}

func (m *Metadata) initDefaults() {
	for _, field := range m.defaultFieldsArray {
		m.form.AddFormItem(field)
	}
	m.form.AddFormItem(m.archived)
}

func (m *Metadata) initButtons() {
	m.form.AddButton("Edit", m.toggleEdit)
}

func (m *Metadata) initCustomFields() {
	if len(*m.bookmark.Metadata) == 0 {
		return
	}

	for _, key := range *m.bookmark.MetadataKeys {

		(*m.customFields)[key] = tview.NewInputField().SetLabel(key).SetText((*m.bookmark.Metadata)[key]).
			SetAcceptanceFunc(m.editEnabled)
		m.form.AddFormItem((*m.customFields)[key])
	}
}

func (m *Metadata) editEnabled(text string, last rune) bool {
	return m.enableEdit
}

func (m *Metadata) cancel() {
	m.setData(m.bookmark)
	m.tmpBookmark = nil
	m.exitEdit()
	m.form.ClearButtons()
	m.initButtons()
}

func (m *Metadata) save() {
	m.tmpBookmark = &models.Bookmark{
		Id:           m.bookmark.Id,
		Name:         m.defaultFields[metadataName].GetText(),
		LowerName:    "",
		Description:  m.defaultFields[metadataDescription].GetText(),
		Content:      m.defaultFields[metadataLink].GetText(),
		Project:      m.defaultFields[metadataProject].GetText(),
		CreatedAt:    m.bookmark.CreatedAt,
		UpdatedAt:    time.Now(),
		Archived:     m.archived.IsChecked(),
		Tags:         nil,
		Metadata:     m.bookmark.Metadata,
		MetadataKeys: m.bookmark.MetadataKeys,
	}

	m.tmpBookmark.LowerName = strings.ToLower(m.tmpBookmark.Name)
	tags := m.defaultFields[metadataTags].GetText()
	if tags != "" {
		tags = strings.Replace(tags, " ", "", -1)
		m.tmpBookmark.Tags = strings.Split(tags, ",")
	}

	for _, key := range *m.customKeys {
		value := (*m.customFields)[key].GetText()
		(*m.tmpBookmark.Metadata)[key] = value
	}

	ok := m.doneFunc(true, m.tmpBookmark)
	if ok {
		m.bookmark = m.tmpBookmark
		m.setFields(m.bookmark)
	}
	m.enableEdit = false
	m.form.SetFieldBackgroundColor(config.Configuration.Colors.Metadata.TextBackground)
	m.form.ClearButtons()
	m.initButtons()
}

func (m *Metadata) exitEdit() {
	m.enableEdit = false
	m.form.SetFieldBackgroundColor(config.Configuration.Colors.Metadata.TextBackground)
}

func (m *Metadata) getTitle() {
	url := m.defaultFields["Link"].GetText()
	if url == "" {
		return
	}
	metadata, err := external.GetPageMetadata(url)
	if err != nil {
		logrus.Errorf("get site title: %v", err)
	} else {
		(*m.customFields)["Title"].SetText(metadata.Title)
	}
}
