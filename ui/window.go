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
	"tryffel.net/go/twidgets"
	"tryffel.net/pkg/bookmarker/config"
	"tryffel.net/pkg/bookmarker/storage/models"
	"tryffel.net/pkg/bookmarker/ui/modals"
)

var navBarLabels = make([]string, 0)
var navBarShortucts = make([]tcell.Key, 0)

type Window struct {
	layout   *twidgets.ModalLayout
	grid     *tview.Grid
	gridAxis []int
	gridSize int

	navBar    *twidgets.NavBar
	project   *Projects
	tags      *Tags
	bookmarks *BookmarkTable

	help         *modals.Help
	bookmarkForm *modals.BookmarkForm

	createFunc func(bookmark *models.Bookmark)
}

func (w *Window) Draw(screen tcell.Screen) {
	w.grid.Draw(screen)
}

func (w *Window) GetRect() (int, int, int, int) {
	return w.grid.GetRect()
}

func (w *Window) SetRect(x, y, width, height int) {
	w.grid.SetRect(x, y, width, height)
}

func (w *Window) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return w.grid.InputHandler()
}

func (w *Window) Focus(delegate func(p tview.Primitive)) {
	w.grid.Focus(delegate)
}

func (w *Window) Blur() {
	w.grid.Blur()
}

func (w *Window) GetFocusable() tview.Focusable {
	return w.grid.GetFocusable()
}

func NewWindow(colors config.Colors, shortcuts *config.Shortcuts) *Window {
	w := &Window{
		layout:    twidgets.NewModalLayout(),
		grid:      tview.NewGrid(),
		project:   NewProjects(),
		tags:      NewTags(),
		bookmarks: NewBookmarkTable(),
		help:      modals.NewHelp(),
	}

	w.bookmarkForm = modals.NewBookmarkForm(w.createBookmark)
	w.grid.SetBackgroundColor(colors.Background)

	w.gridSize = 6
	w.grid.SetRows(1, -1)
	w.grid.SetColumns(-1)
	w.grid.SetMinSize(2, 2)

	col := colors.NavBar.ToNavBar()
	w.navBar = twidgets.NewNavBar(col, w.navBarClicked)
	navBarLabels = []string{"Help", "New Bookmark", "Menu", "Quit"}

	sc := shortcuts.NavBar
	navBarShortucts = []tcell.Key{sc.Help, sc.NewBookmark, sc.Menu, sc.Quit}

	for i, v := range navBarLabels {
		btn := tview.NewButton(v)
		w.navBar.AddButton(btn, navBarShortucts[i])
	}

	w.grid.AddItem(w.navBar, 0, 0, 1, 1, 1, 10, false)
	w.grid.AddItem(w.layout, 1, 0, 1, 1, 4, 4, true)

	w.layout.Grid().AddItem(w.project, 0, 0, 3, 2, 5, 5, false)
	w.layout.Grid().AddItem(w.tags, 3, 0, 3, 2, 5, 5, false)
	w.layout.Grid().AddItem(w.bookmarks, 0, 2, 6, 4, 10, 10, true)
	return w
}

func (w *Window) navBarClicked(label string) {
	logrus.Info("User pressed: ", label)

}

func (w *Window) createBookmark(b *models.Bookmark) {
	if w.createFunc != nil {
		w.createFunc(b)

	}
}
