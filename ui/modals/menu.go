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
	"tryffel.net/go/bookmarker/config"
)

type MenuAction int

const (
	MenuActionNone MenuAction = iota
	MenuActionImport
	MenuActionExport
	MenuActionModify
)

//Menu provides modal to perform multiple actions
type Menu struct {
	*tview.List
	doneFunc func(action MenuAction)
}

func NewMenu() *Menu {
	m := &Menu{
		List: tview.NewList(),
	}

	colors := config.Configuration.Colors.BookmarkForm
	m.SetBackgroundColor(colors.Background)
	m.SetBorder(true)
	m.SetTitle("Menu")
	m.SetBorderColor(config.Configuration.Colors.Border)
	m.SetMainTextColor(colors.Text)
	m.SetSecondaryTextColor(colors.Label)
	//m.SetSelectedTextColor(colors.TextSelected)
	m.SetSelectedBackgroundColor(colors.TextSelected)

	m.AddItem("Import bookmarks", "Import from bookmarks.html file", 'i', m.doImport)
	m.AddItem("Export bookmarks (not implemented)", "Export into bookmarks.html file", 'e', m.doExport)
	m.AddItem("Bulk Modify", "Modify multiple bookmarks with given filter", 'm', m.doModify)

	return m
}

func (m *Menu) SetDoneFunc(doneFunc func()) {
}

func (m *Menu) SetActionFunc(doneFunc func(action MenuAction)) {
	m.doneFunc = doneFunc
}

func (m *Menu) SetVisible(visible bool) {
}

func (m *Menu) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			if m.doneFunc != nil {
				m.doneFunc(MenuActionNone)
			}
		} else {
			m.List.InputHandler()(event, setFocus)
		}
	}
}

func (m *Menu) doImport() {
	if (m.doneFunc) != nil {
		m.doneFunc(MenuActionImport)
	}
}

func (m *Menu) doExport() {
	if (m.doneFunc) != nil {
		m.doneFunc(MenuActionExport)
	}
}

func (m *Menu) doModify() {
	if (m.doneFunc) != nil {
		m.doneFunc(MenuActionModify)
	}
}
