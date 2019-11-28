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
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"os"
	"time"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/external"
	"tryffel.net/go/bookmarker/storage"
	"tryffel.net/go/bookmarker/storage/models"
	"tryffel.net/go/bookmarker/ui/modals"
	"tryffel.net/go/twidgets"
)

var navBarLabels = make([]string, 0)
var navBarShortucts = make([]tcell.Key, 0)

// Navigating over widgets with tab in this order

type Window struct {
	app *tview.Application
	db  *storage.Database

	layout   *twidgets.ModalLayout
	grid     *tview.Grid
	gridAxis []int
	gridSize int

	navBar     *twidgets.NavBar
	project    *Projects
	tags       *Tags
	bookmarks  *BookmarkTable
	metadata   *Metadata
	search     *Search
	menu       *modals.Menu
	importForm *modals.ImportForm
	searchOpen bool

	help         *modals.Help
	bookmarkForm *modals.BookmarkForm

	hasModal  bool
	modal     twidgets.Modal
	lastFocus tview.Primitive

	tabWidgetCount int
	tabWidgets     []tview.Primitive
	createFunc     func(bookmark *models.Bookmark)

	metadataOpen bool

	filter *storage.Filter
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
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		//key := event.Key()
		//if key == tcell.KeyCtrlSpace {
		//	w.openMetadata()
		//} else {
		w.grid.InputHandler()(event, setFocus)
		//}
	}
}

func (w *Window) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	navbar := config.Configuration.Shortcuts.NavBar
	key := event.Key()
	switch key {
	case navbar.Help:
		if !w.hasModal {

			stats, err := w.db.GetStatistics()
			if err != nil {
				logrus.Errorf("Get statistics: &v", err)
			}

			w.addModal(w.help, twidgets.ModalSizeMedium)
			w.help.Update(stats)
		}
	case navbar.NewBookmark:
		w.addModal(w.bookmarkForm, twidgets.ModalSizeMedium)
	case navbar.OpenBrowser:
		index, _ := w.bookmarks.table.GetSelection()
		bookmark := w.bookmarks.items[index-1]

		err := external.OpenUrlInBrowser(bookmark.Content)
		if err != nil {
			logrus.Errorf("Open link in browser: %v", err)
		}
	case navbar.Menu:
		w.addModal(w.menu, twidgets.ModalSizeMedium)
	case navbar.Quit:
		w.quit()
	case tcell.KeyEscape:
		if w.hasModal {
			w.layout.RemoveModal(w.modal)
			w.app.SetFocus(w.lastFocus)
			w.lastFocus = nil
			w.modal = nil
			w.hasModal = false
		} else if w.metadataOpen {
			w.closeMetadata(false, nil)
		} else if w.searchOpen {
			w.app.SetFocus(w.lastFocus)
			w.lastFocus = nil
			w.searchOpen = false
		}
	case tcell.KeyCtrlSpace:
		//case tcell.KeyEnter:
		if !w.metadataOpen || !w.hasModal {
			w.openMetadata()
		}

	case tcell.KeyCtrlD:
		w.closeMetadata(false, nil)
		w.lastFocus = w.app.GetFocus()
		w.app.SetFocus(w.search)
		w.searchOpen = true

	case tcell.KeyTAB:
		if !w.metadataOpen && !w.hasModal {
			w.nextWidget()
		} else {
			return event
		}
	default:
		return event
	}
	return nil

}

func (w *Window) nextWidget() {
	next := w.tabWidgetCount + 1
	if next >= len(w.tabWidgets) {
		next = 0
	}

	w.app.SetFocus(w.tabWidgets[next])
	w.tabWidgetCount = next
}

func (w *Window) addModal(modal twidgets.Modal, size twidgets.ModalSize) {
	if !w.hasModal {
		w.layout.AddDynamicModal(modal, size)

		w.lastFocus = w.app.GetFocus()
		w.app.SetFocus(modal)
		w.modal = modal
		w.hasModal = true
	}
}

func (w *Window) Focus(delegate func(p tview.Primitive)) {
	w.grid.Focus(delegate)
}

func (w *Window) Blur() {
	w.grid.Blur()
}

func (w *Window) GetFocusable() tview.Focusable {
	return w.layout.GetFocusable()
}

func NewWindow(colors config.Colors, shortcuts *config.Shortcuts, db *storage.Database) *Window {
	w := &Window{
		app:        tview.NewApplication(),
		db:         db,
		layout:     twidgets.NewModalLayout(),
		grid:       tview.NewGrid(),
		project:    NewProjects(),
		tags:       NewTags(),
		help:       modals.NewHelp(),
		importForm: modals.NewImportForm(),
	}

	w.app.SetRoot(w, true)
	w.app.SetInputCapture(w.inputCapture)

	w.layout.SetGridYSize([]int{3, -1, -1, -1, -1, -1, -1, -1, -1, 3})
	w.bookmarks = NewBookmarkTable(w.openBookmark)
	w.bookmarks.SetDeleteFunc(w.deleteBookmark)
	w.metadata = NewMetadata(w.closeMetadata)
	w.metadata.SetSearchFunc(w.autoComplete)

	w.bookmarkForm = modals.NewBookmarkForm(w.createBookmark)
	w.bookmarkForm.SetSearchFunc(w.autoComplete)
	w.grid.SetBackgroundColor(colors.Background)
	w.search = NewSearch(w.Search)
	w.project.SetSelectFunc(w.FilterByProject)
	w.menu = modals.NewMenu()
	w.menu.SetActionFunc(w.menuAction)
	w.importForm.SetCreateFunc(w.doImport)
	w.modify = modals.NewModify(w.modifyBookmark)

	w.gridSize = 6
	w.grid.SetRows(1, -1)
	w.grid.SetColumns(-1)
	w.grid.SetMinSize(1, 2)

	col := colors.NavBar.ToNavBar()

	//w.metadata = NewMetadata(w.closeMetadata)
	w.navBar = twidgets.NewNavBar(col, w.navBarClicked)
	navBarLabels = []string{"Help", "New Bookmark", "Open link", "Menu", "Quit"}

	sc := shortcuts.NavBar
	navBarShortucts = []tcell.Key{sc.Help, sc.NewBookmark, sc.OpenBrowser, sc.Menu, sc.Quit}

	for i, v := range navBarLabels {
		btn := tview.NewButton(v)
		w.navBar.AddButton(btn, navBarShortucts[i])
	}

	w.grid.AddItem(w.navBar, 0, 0, 1, 1, 1, 10, false)
	w.grid.AddItem(w.layout, 1, 0, 1, 1, 4, 4, true)

	w.tabWidgets = append(w.tabWidgets, w.bookmarks)
	w.tabWidgets = append(w.tabWidgets, w.project)
	w.tabWidgets = append(w.tabWidgets, w.tags)

	w.initDefaultLayout()
	w.app.SetFocus(w.bookmarks)
	return w
}

func (w *Window) navBarClicked(label string) {
	logrus.Info("User pressed: ", label)

}

func (w *Window) closeMetadata(save bool, bookmark *models.Bookmark) bool {
	if !w.metadataOpen {
		return false
	}
	if save {
		err := w.db.UpdateBookmark(bookmark)

		if err != nil {
			logrus.Errorf("Failed to update bookmark %d %s: %v", bookmark.Id, bookmark.Name, err)
			// TODO: show modal for error?
			return false
		} else {
			return true
		}

	}

	if !save {
		w.initDefaultLayout()
	}
	w.app.SetFocus(w.lastFocus)
	w.lastFocus = nil
	w.metadataOpen = false
	return false
}

func (w *Window) initDefaultLayout() {
	w.layout.Grid().Clear()

	w.layout.Grid().AddItem(w.project, 0, 0, 7, 1, 5, 5, false)
	w.layout.Grid().AddItem(w.tags, 7, 0, 3, 1, 5, 5, false)
	w.layout.Grid().AddItem(w.bookmarks, 0, 1, 9, 9, 10, 10, true)
	w.layout.Grid().AddItem(w.search, 9, 1, 1, 9, 1, 10, false)
}

func (w *Window) openBookmark(b *models.Bookmark) {
	w.openMetadata()
	w.metadata.setData(b)
}

func (w *Window) openMetadata() {
	w.lastFocus = w.app.GetFocus()

	//w.grid.Blur()
	//w.metadata.Focus(func(p tview.Primitive){})
	w.layout.Grid().RemoveItem(w.bookmarks)
	w.layout.Grid().RemoveItem(w.project)
	w.layout.Grid().RemoveItem(w.tags)
	w.layout.Grid().RemoveItem(w.search)

	w.layout.Grid().AddItem(w.bookmarks, 0, 0, 9, 7, 10, 10, false)
	w.layout.Grid().AddItem(w.search, 9, 0, 1, 7, 1, 10, false)
	w.layout.Grid().AddItem(w.metadata, 0, 7, 10, 3, 10, 10, true)

	index, _ := w.bookmarks.table.GetSelection()
	bookmark := w.bookmarks.items[index-1]

	w.app.QueueUpdateDraw(func() {
		err := w.db.GetBookmarkMetadata(bookmark)
		if err != nil {
			logrus.Errorf("Get metadata: %v", err)
		}
		w.metadata.setData(bookmark)
	})
	w.metadataOpen = true
	w.app.SetFocus(w.metadata)
}

func (w *Window) createBookmark(bookmark *models.Bookmark) {
	logrus.Debugf("Create new bookmark: ", bookmark)

	err := w.db.NewBookmark(bookmark)
	if err != nil {
		logrus.Error("Failed to create bookmark: ", err)
	} else {
		w.bookmarkForm.Clear()
		bookmarks, err := w.db.GetAllBookmarks()
		if err != nil {
			return
		}
		w.bookmarks.SetData(bookmarks)
		if w.hasModal {
			w.layout.RemoveModal(w.modal)
			w.app.SetFocus(w.lastFocus)
			w.lastFocus = nil
			w.modal = nil
			w.hasModal = false
		}
	}
}

func (w *Window) Search(text string) {
	var err error
	w.filter, err = storage.NewFilter(text)
	if err != nil || w.filter.IsPlainQuery() {
		logrus.Errorf("Failed to parse query: %v", err)
		bookmarks, err := w.db.SearchBookmarks(text)
		if err != nil {
			logrus.Errorf("Search bookmarks: %v", err)
			return
		}
		w.bookmarks.SetData(bookmarks)
		w.bookmarks.ResetCursor()
	} else {
		bookmarks, err := w.db.SearchBookmarksFilter(w.filter)
		if err != nil {
			logrus.Errorf("Search bookmarks: %v", err)
			return
		}

		w.bookmarks.SetData(bookmarks)
		w.bookmarks.ResetCursor()

	}
}

func (w *Window) FilterByProject(project *models.Project) {
	if project == nil {
		bookmarks, err := w.db.GetAllBookmarks()
		if err != nil {
			logrus.Error("Get all bookmarks: %v", err)
		} else {
			w.bookmarks.SetData(bookmarks)
			w.bookmarks.ResetCursor()
		}
	} else {
		name := ""
		strict := true
		name = project.FullName()
		logrus.Debug("Filtering with projects: ", name)
		if project.Parent != nil || len(project.Children) > 0 {
			strict = false
		}
		bookmarks, err := w.db.GetProjectBookmarks(name, strict)
		if err != nil {

			logrus.Error("Get bookmarks by project: %v", err)
		} else {
			w.bookmarks.SetData(bookmarks)
			w.bookmarks.ResetCursor()
		}
	}
}

func (w *Window) menuAction(action modals.MenuAction) {
	switch action {
	case modals.MenuActionNone:
	case modals.MenuActionImport:
		w.layout.RemoveModal(w.modal)
		w.hasModal = false
		w.addModal(w.importForm, twidgets.ModalSizeMedium)
	case modals.MenuActionModify:
		w.layout.RemoveModal(w.modal)
		w.hasModal = false
		w.addModal(w.modify, twidgets.ModalSizeMedium)
	}
}

func (w *Window) doImport(data *modals.ImportData) {
	logrus.Info("User wants to import: ", data)

	ok := false
	msg := ""
	count := 0

	file, err := os.Open(data.File)
	defer file.Close()
	if err != nil {
		logrus.Error(err)
		msg = fmt.Errorf("failed to open file: %v", err).Error()
	} else {
		start := time.Now()
		bookmarks, err := external.ImportBookmarksHtml(file, data.MapFoldersProjects)
		if err != nil {
			logrus.Error(err)
			msg = fmt.Errorf("parse bookmarks.html: %v", err).Error()
		} else {
			err = w.db.NewBookmarks(bookmarks, data.Tags)
			took := time.Since(start)
			if err != nil {
				logrus.Error("Batch import and create bookmarks: %v", err)
				msg = fmt.Errorf("save new bookmarks: %v", err).Error()
			} else {
				logrus.Infof("Imported %d bookmarks in %d ms", len(bookmarks), took.Milliseconds())
				ok = true
				msg = fmt.Sprintf("Took %d ms", took.Milliseconds())
				count = len(bookmarks)
			}
		}
	}
	w.importForm.SetDoneFunc(w.closeImport)
	w.importForm.ImportDone(count, msg, ok)
}

func (w *Window) closeImport() {
	w.importForm.Reset()
	w.closeModal()
}

func (w *Window) closeModal() {
	if w.hasModal {
		w.layout.RemoveModal(w.modal)
		w.app.SetFocus(w.lastFocus)
		w.lastFocus = nil
		w.modal = nil
		w.hasModal = false
	}
}

func (w *Window) deleteBookmark(bookmark *models.Bookmark) {
	doneFunc := func(del bool) {
		if del {
			err := w.db.DeleteBookmark(bookmark)
			if err != nil {
				logrus.Errorf("Delete bookmark: %v", err)
			}
			w.RefreshBookmarks()
		}
		w.closeModal()
	}

	del := modals.NewDeleteBookmark(doneFunc, bookmark)
	w.addModal(del, twidgets.ModalSizeSmall)
}

func (w *Window) RefreshBookmarks() {
	bookmarks, err := w.db.GetAllBookmarks()
	if err != nil {
		return
	}
	w.bookmarks.SetData(bookmarks)
}

func (w *Window) Run() error {
	bookmarks, _ := w.db.GetAllBookmarks()
	projects, _ := w.db.GetProjects("", false)
	tags, _ := w.db.GetTags()

	w.bookmarks.SetData(bookmarks)
	w.tags.SetData(tags)
	w.project.SetData(projects)

	return w.app.Run()
}

func (w *Window) quit() {
	w.app.Stop()
}

func (w *Window) modifyBookmark(filter *storage.Filter, modifier *storage.Modifier) (int, error) {
	return w.db.BulkModify(filter, modifier)
}

func (w *Window) autoComplete(key, value string) ([]string, error) {
	if config.Configuration.AutoComplete {
		return w.db.SearchKeyValue(key, value)
	} else {
		return nil, nil
	}

}
