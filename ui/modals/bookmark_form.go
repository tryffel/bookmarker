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
	"strings"
	"time"
	"tryffel.net/pkg/bookmarker/config"
	"tryffel.net/pkg/bookmarker/storage/models"
)

type BookmarkForm struct {
	form     *tview.Form
	doneFunc func()
	formFunc func(bookmark *models.Bookmark)
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

	b.form.AddInputField("Name", "bookmark", 0, nil, nil)
	b.form.AddInputField("Description", "short description", 0, nil, nil)
	b.form.AddInputField("Link", "https://..", 0, nil, nil)
	b.form.AddInputField("Project", "e.g", 0, nil, nil)

	b.form.AddButton("Create", b.create)
	b.form.AddButton("Cancel", b.doneFunc)

	return b
}

func (n *BookmarkForm) SetDoneFunc(doneFunc func()) {
	n.doneFunc = doneFunc
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
		Name:        n.form.GetFormItemByLabel("Name").(*tview.InputField).GetText(),
		LowerName:   "",
		Description: n.form.GetFormItemByLabel("Description").(*tview.InputField).GetText(),
		Content:     n.form.GetFormItemByLabel("Link").(*tview.InputField).GetText(),
		Project:     n.form.GetFormItemByLabel("Project").(*tview.InputField).GetText(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bookmark.LowerName = strings.ToLower(bookmark.Name)
	n.formFunc(bookmark)
}
