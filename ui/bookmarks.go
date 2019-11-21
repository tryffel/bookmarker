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
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/storage/models"
)

type BookmarkTable struct {
	table        *tview.Table
	items        []*models.Bookmark
	hasFocus     bool
	metadataFunc func(bookmark *models.Bookmark)
	deleteFunc   func(bookmark *models.Bookmark)
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

		} else {
			b.table.InputHandler()(event, setFocus)
		}
	}
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

	b.table.Clear()
	b.table.SetCell(0, 0, tableHeaderCell("#"))
	b.table.SetCell(0, 1, tableHeaderCell("Name"))
	b.table.SetCell(0, 2, tableHeaderCell("Description"))
	b.table.SetCell(0, 3, tableHeaderCell("Project"))
	b.table.SetCell(0, 4, tableHeaderCell("Link"))
	b.table.SetCell(0, 5, tableHeaderCell("Tags"))
	b.table.SetCell(0, 6, tableHeaderCell("Added at"))
	b.table.SetOffset(1, 0)

	for i, v := range data {
		b.table.SetCell(i+1, 0, tableCell(fmt.Sprint(i+1)))
		b.table.SetCell(i+1, 1, tableCell(v.Name).SetMaxWidth(25).SetExpansion(1))
		b.table.SetCell(i+1, 2, tableCell(v.Description).SetMaxWidth(35).SetExpansion(3))
		b.table.SetCell(i+1, 3, tableCell(v.Project).SetMaxWidth(20).SetExpansion(1))
		b.table.SetCell(i+1, 4, tableCell(v.ContentDomain()).SetMaxWidth(20).SetExpansion(1))
		tags := tableCell(v.TagsString(true)).SetMaxWidth(13).SetExpansion(1)
		tags.SetTextColor(config.Configuration.Colors.Tags.Text)
		b.table.SetCell(i+1, 5, tags)
		b.table.SetCell(i+1, 6, tableCell(ShortTimeSince(v.CreatedAt)).SetMaxWidth(15))
	}

}

func (b *BookmarkTable) ResetCursor() {
	b.table.Select(1, 0)

}

func (b *BookmarkTable) SetDeleteFunc(delete func(bookmark *models.Bookmark)) {
	b.deleteFunc = delete
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
		table:        tview.NewTable(),
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
