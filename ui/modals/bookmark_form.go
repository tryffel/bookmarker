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
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/external"
	"tryffel.net/go/bookmarker/storage/models"
)

type BookmarkForm struct {
	form       *tview.Form
	doneFunc   func()
	formFunc   func(bookmark *models.Bookmark)
	searchFunc func(key, value string) ([]string, error)

	nameField        *tview.InputField
	descriptionField *tview.InputField
	linkField        *tview.InputField
	projectField     *tview.InputField
	tagsField        *tview.InputField
}

func NewBookmarkForm(createFunc func(bookmark *models.Bookmark)) *BookmarkForm {
	b := &BookmarkForm{
		form:     tview.NewForm(),
		doneFunc: nil,
		formFunc: createFunc,
	}

	colors := config.Configuration.Colors.BookmarkForm

	b.form.SetTitle("New bookmark")
	b.form.SetTitleColor(colors.Text)

	b.form.SetBorder(true)
	b.form.SetBorderColor(config.Configuration.Colors.Border)
	b.form.SetBackgroundColor(colors.Background)
	b.form.SetLabelColor(colors.Label)
	b.form.SetFieldBackgroundColor(colors.TextBackground)
	b.form.SetFieldTextColor(colors.Text)

	b.nameField = tview.NewInputField().SetLabel("Name").SetPlaceholder("bookmark")
	b.descriptionField = tview.NewInputField().SetLabel("Description").SetPlaceholder("my bookmark")
	b.linkField = tview.NewInputField().SetLabel("Link").SetPlaceholder("https://...")
	b.projectField = tview.NewInputField().SetLabel("Project").SetPlaceholder("bookmarks.a").
		SetAutocompleteFunc(b.search("Project"))
	b.tagsField = tview.NewInputField().SetLabel("Tags").SetPlaceholder("a,b")

	b.nameField.SetPlaceholderTextColor(colors.TextPlaceHolder)
	b.descriptionField.SetPlaceholderTextColor(colors.TextPlaceHolder)
	b.linkField.SetPlaceholderTextColor(colors.TextPlaceHolder)
	b.projectField.SetPlaceholderTextColor(colors.TextPlaceHolder)
	b.tagsField.SetPlaceholderTextColor(colors.TextPlaceHolder)

	b.initForm()
	return b
}

func (n *BookmarkForm) SetDoneFunc(doneFunc func()) {
	n.doneFunc = doneFunc
}

func (n *BookmarkForm) SetSearchFunc(search func(key, value string) ([]string, error)) {
	n.searchFunc = search
}

func (n *BookmarkForm) Draw(screen tcell.Screen) {
	n.form.Draw(screen)
}

func (n *BookmarkForm) GetRect() (int, int, int, int) {
	return n.form.GetRect()
}

func (n *BookmarkForm) SetRect(x, y, width, height int) {
	n.form.SetRect(x, y, width, height)
}

func (n *BookmarkForm) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return n.form.InputHandler()
}

func (n *BookmarkForm) Focus(delegate func(p tview.Primitive)) {
	n.form.Focus(delegate)
}

func (n *BookmarkForm) Blur() {
	n.form.Blur()
}

func (n *BookmarkForm) GetFocusable() tview.Focusable {
	return n.form.GetFocusable()
}

func (n *BookmarkForm) SetVisible(visible bool) {
}

func (n *BookmarkForm) create() {
	bookmark := &models.Bookmark{
		Id:          0,
		Name:        n.nameField.GetText(),
		LowerName:   "",
		Description: n.descriptionField.GetText(),
		Content:     n.linkField.GetText(),
		Project:     n.projectField.GetText(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bookmark.FillDefaultMetadata()
	bookmark.LowerName = strings.ToLower(bookmark.Name)
	tags := n.tagsField.GetText()
	if tags != "" {
		tags = strings.Replace(tags, " ", "", -1)
		bookmark.Tags = strings.Split(tags, ",")
	}

	for _, key := range models.DefaulMetadata {
		item := n.form.GetFormItemByLabel(key)
		if item != nil {
			text := item.(*tview.InputField).GetText()
			(*bookmark.Metadata)[key] = text
		}
	}

	n.formFunc(bookmark)
}

func (n *BookmarkForm) Clear() {
	n.form.Clear(true)
	n.nameField.SetText("")
	n.descriptionField.SetText("")
	n.linkField.SetText("")
	n.projectField.SetText("")
	n.tagsField.SetText("")
	n.initForm()
}

func (n *BookmarkForm) initForm() {
	n.form.AddFormItem(n.nameField)
	n.form.AddFormItem(n.descriptionField)
	n.form.AddFormItem(n.linkField)
	n.form.AddFormItem(n.projectField)
	n.form.AddFormItem(n.tagsField)
	custom := models.DefaulMetadata
	for _, v := range custom {
		field := tview.NewInputField().SetLabel(v).SetAutocompleteFunc(n.search(v))
		n.form.AddFormItem(field)
		//n.form.AddInputField(v, "", 0, nil, nil).
	}

	n.form.AddButton("Create", n.create)
	n.form.AddButton("Cancel", n.doneFunc)
	n.form.AddButton("Get title", n.getTitle)
}

func (n *BookmarkForm) getTitle() {
	if n.linkField.GetText() == "" {
		return
	}
	metadata, err := external.GetPageMetadata(n.linkField.GetText())
	if err != nil {
		logrus.Errorf("get site title: %v", err)
	} else {
		n.form.GetFormItemByLabel("Title").(*tview.InputField).SetText(metadata.Title)
	}
}

func (n *BookmarkForm) wrapSearch(field string) func(string) []string {
	return n.search(field)
}

//do searching a.k.a completion
func (n *BookmarkForm) search(field string) func(string) []string {
	return func(text string) []string {
		if text == "" || field == "" || n.searchFunc == nil {
			return nil
		}

		results, err := n.searchFunc(field, text)
		if err != nil {
			logrus.Errorf("autocomplete field: %v", err)
			return nil
		} else {
			return results
		}
	}
}
