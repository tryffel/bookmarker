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
	"tryffel.net/pkg/bookmarker/config"
	"tryffel.net/pkg/bookmarker/storage/models"
)

type BookmarkTable struct {
	table    *tview.Table
	items    []*models.Bookmark
	hasFocus bool
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
	return b.table.InputHandler()
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
	b.items = data

	b.table.Clear()
	b.table.SetCell(0, 0, tableHeaderCell("#"))
	b.table.SetCell(0, 1, tableHeaderCell("Name"))
	b.table.SetCell(0, 2, tableHeaderCell("Description"))
	b.table.SetCell(0, 3, tableHeaderCell("Link"))
	b.table.SetCell(0, 4, tableHeaderCell("Added at"))
	b.table.SetOffset(1, 0)

	for i, v := range data {
		b.table.SetCell(i+1, 0, tableCell(fmt.Sprint(i+1)))
		b.table.SetCell(i+1, 1, tableCell(v.Name))
		b.table.SetCell(i+1, 2, tableCell(v.Description))
		b.table.SetCell(i+1, 3, tableCell(v.Content))
		b.table.SetCell(i+1, 4, tableCell(v.CreatedAt.Format("2006-01-02 15:04")))
	}

	b.table.Select(1, 0)

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

func NewBookmarkTable() *BookmarkTable {
	b := &BookmarkTable{
		table: tview.NewTable(),
		items: []*models.Bookmark{},
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
