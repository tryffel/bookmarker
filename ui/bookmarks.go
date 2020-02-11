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
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/storage/models"
	"tryffel.net/go/twidgets"
)

type BookmarkTable struct {
	table        *twidgets.Table
	items        []*models.Bookmark
	hasFocus     bool
	metadataFunc func(bookmark *models.Bookmark)
	deleteFunc   func(bookmark *models.Bookmark)
	sortFunc     func(column string, sort twidgets.Sort)
}

func (b *BookmarkTable) Draw(screen tcell.Screen) {
	b.table.Draw(screen)
}

func (b *BookmarkTable) GetRect() (int, int, int, int) {
	return b.table.GetRect()
}

func (b *BookmarkTable) SetRect(x, y, width, height int) {
	b.table.SetRect(x, y, width, height)
}

func (b *BookmarkTable) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyCtrlSpace {
			selected, _ := b.table.GetSelection()
			bookmark := b.items[selected]
			b.metadataFunc(bookmark)
		} else if key == tcell.KeyDelete {
			if b.deleteFunc != nil {
				index, _ := b.table.GetSelection()
				bookmark := b.items[index-1]
				b.deleteFunc(bookmark)
			}
		} else if event.Rune() == 'n' {
			b.moveCursor(10)
		} else if event.Rune() == 'm' {
			b.moveCursor(-10)
		} else {
			b.table.InputHandler()(event, setFocus)
		}
	}
}

func (b *BookmarkTable) moveCursor(n int) {
	index, col := b.table.GetSelection()
	result := index + n
	if result >= len(b.items) {
		result = len(b.items) - 1
	} else if result <= 0 {
		result = 1
	}
	b.table.Select(result, col)
}

func (b *BookmarkTable) Focus(delegate func(p tview.Primitive)) {
	b.hasFocus = true
	b.table.SetBorderColor(config.Configuration.Colors.BorderFocus)
	b.table.Focus(delegate)
}

func (b *BookmarkTable) Blur() {
	b.hasFocus = false
	b.table.SetBorderColor(config.Configuration.Colors.Border)
	b.table.Blur()
}

func (b *BookmarkTable) GetFocusable() tview.Focusable {
	return b.table.GetFocusable()
}

func (b *BookmarkTable) SetData(data []*models.Bookmark) {
	if data == nil {
		return
	}
	b.items = data

	b.table.Clear(false)
	for i, v := range data {
		row := []string{
			v.Name,
			v.Description,
			v.Project,
			v.ContentDomain(),
			v.TagsString(true),
			ShortTimeSince(v.CreatedAt),
		}

		b.table.AddRow(i, row...)
	}
	if len(b.items) > 0 {
		b.table.Select(1, 0)
	}
}

func (b *BookmarkTable) ResetCursor() {
	b.table.Select(1, 0)

}

func (b *BookmarkTable) SetDeleteFunc(delete func(bookmark *models.Bookmark)) {
	b.deleteFunc = delete
}

func (b *BookmarkTable) SetSortFunc(sort func(column string, sort twidgets.Sort)) {
	b.sortFunc = sort
}

func tableCell(text string) *tview.TableCell {
	c := tview.NewTableCell(text)
	c.SetTextColor(config.Configuration.Colors.Bookmarks.Text)
	c.SetAlign(tview.AlignLeft)
	return c
}

func tableHeaderCell(text string) *tview.TableCell {
	c := tview.NewTableCell(text)
	c.SetTextColor(config.Configuration.Colors.Bookmarks.HeaderText)
	c.SetAlign(tview.AlignLeft)
	c.SetSelectable(false)
	return c
}

func NewBookmarkTable(openMetadata func(bookmark *models.Bookmark)) *BookmarkTable {
	b := &BookmarkTable{
		table:        twidgets.NewTable(),
		items:        []*models.Bookmark{},
		metadataFunc: openMetadata,
	}

	colors := config.Configuration.Colors.Bookmarks
	b.table.SetBackgroundColor(colors.Background)
	b.table.SetBorder(true)
	b.table.SetBorders(false)
	b.table.SetBorderColor(config.Configuration.Colors.Border)
	b.table.SetSelectedStyle(colors.TextSelected, colors.BackgroundSelected, 0)
	b.table.SetSelectable(true, false)
	b.table.SetFixed(1, 10)

	b.table.SetAddCellFunc(b.addCell)
	b.table.SetShowIndex(true)
	b.table.SetColumns([]string{"Name", "Description", "Project", "Link", "Tags", "Added at"})
	b.table.SetColumnWidths([]int{3, 25, 35, 20, 10, 15, 10})
	b.table.SetColumnExpansions([]int{0, 1, 3, 1, 1, 1, 1})
	b.table.SetSort(0, twidgets.SortAsc)
	b.table.SetSortFunc(b.sort)
	return b
}

func (b *BookmarkTable) GetSelection() *models.Bookmark {
	index, _ := b.table.GetSelection()
	if b.items == nil {
		return nil
	}
	if index > len(b.items) {
		return nil
	}
	return b.items[index-1]
}

func (b *BookmarkTable) addCell(cell *tview.TableCell, header bool, row int) {
	if header {
		cell.SetTextColor(config.Configuration.Colors.Bookmarks.HeaderText)
		cell.SetAlign(tview.AlignLeft)
	} else {
		cell.SetTextColor(config.Configuration.Colors.Bookmarks.Text)
		if row%2 == 1 {
			cell.SetBackgroundColor(config.Configuration.Colors.Bookmarks.Background2nd)
		}
		cell.SetAlign(tview.AlignLeft)
	}
}

func (b *BookmarkTable) sort(column string, sort twidgets.Sort) {
	if b.sortFunc != nil {
		b.sortFunc(column, sort)
	}
}
