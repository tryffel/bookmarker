/*
 *
 *  Copyright 2019 Tero Vierimaa
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/twidgets"
	"tryffel.net/pkg/bookmarker/config"
	"tryffel.net/pkg/bookmarker/storage"
	"tryffel.net/pkg/bookmarker/storage/models"
)

type Application struct {
	app       *tview.Application
	window    *Window
	db        *storage.Database
	hasModal  bool
	modal     twidgets.Modal
	lastFocus tview.Primitive
}

func NewApplication(colors config.Colors, shortcuts *config.Shortcuts, db *storage.Database) *Application {
	a := &Application{
		app:    tview.NewApplication(),
		window: NewWindow(colors, shortcuts),
		db:     db,
	}

	a.app.SetRoot(a.window, true)
	a.app.SetInputCapture(a.inputCapture)
	a.window.createFunc = a.createBookmark
	return a
}

func (a *Application) Run() error {
	err := a.app.Run()
	return err
}

func (a *Application) Initdata() {
	bookmarks, _ := a.db.GetAllBookmarks()
	//projects, _ := a.db.GetProjects("", false)

	tags, _ := a.db.GetTags()

	a.window.bookmarks.SetData(bookmarks)
	a.window.tags.SetData(tags)
	//a.window.project.SetData(projects)

}

func (a *Application) inputCapture(eventKey *tcell.EventKey) *tcell.EventKey {
	navbar := config.Configuration.Shortcuts.NavBar
	key := eventKey.Key()
	switch key {
	case navbar.Menu:
	case navbar.Help:
		if !a.hasModal {
			a.addModal(a.window.help, 10, 40, true)
			a.window.help.Update()
		}
	case navbar.NewBookmark:
		a.addModal(a.window.bookmarkForm, 10, 40, false)
	case tcell.KeyEscape:
		if a.hasModal {
			a.window.layout.RemoveModal(a.modal)
			a.app.SetFocus(a.lastFocus)
			a.lastFocus = nil
			a.modal = nil
			a.hasModal = false
		}
	default:
		return eventKey
	}
	return nil
}

func (a *Application) addModal(modal twidgets.Modal, h, w uint, lockSize bool) {
	if !a.hasModal {
		a.window.layout.AddModal(modal, h, w, lockSize)

		a.lastFocus = a.app.GetFocus()
		a.app.SetFocus(modal)
		a.modal = a.window.bookmarkForm
		a.hasModal = true
	}
}

func (a *Application) createBookmark(bookmark *models.Bookmark) {
	logrus.Info("Got new bookmark: ", bookmark)

	err := a.db.NewBookmark(bookmark)
	if err != nil {
		logrus.Error("Failed to create bookmark: ", err)
	} else {
		bookmarks, err := a.db.GetAllBookmarks()
		if err != nil {
			return
		}
		a.window.bookmarks.SetData(bookmarks)
		if a.hasModal {
			a.window.layout.RemoveModal(a.modal)
			a.app.SetFocus(a.lastFocus)
			a.lastFocus = nil
			a.modal = nil
			a.hasModal = false
		}

	}
}
