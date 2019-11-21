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
)

type Tags struct {
	table         *tview.Table
	tags          *map[string]int
	lastSelectRow int
}

func NewTags() *Tags {
	t := &Tags{table: tview.NewTable()}

	colors := config.Configuration.Colors.Tags
	t.table.SetTitle("Tags")
	t.table.SetTitleColor(config.Configuration.Colors.TextPrimary)
	t.table.SetBackgroundColor(colors.Background)
	t.table.SetBorder(true)
	t.table.SetBorders(false)
	t.table.SetBorderColor(config.Configuration.Colors.Border)
	t.table.SetSelectedStyle(colors.TextSelected, colors.BackgroundSelected, 0)
	t.table.SetSelectable(true, false)
	return t
}

func (t *Tags) Draw(screen tcell.Screen) {
	t.table.Draw(screen)
}

func (t *Tags) GetRect() (int, int, int, int) {
	return t.table.GetRect()
}

func (t *Tags) SetRect(x, y, width, height int) {
	t.table.SetRect(x, y, width, height)
}

func (t *Tags) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.table.InputHandler()
}

func (t *Tags) Focus(delegate func(p tview.Primitive)) {
	t.table.Focus(delegate)
	t.table.SetBorderColor(config.Configuration.Colors.BorderFocus)

	// Only reset selection if there's no previous selection, that is, data has just been updated
	if t.lastSelectRow == -1 {
		t.table.Select(0, 0)
	}
}

func (t *Tags) Blur() {
	t.table.Blur()
	t.table.SetBorderColor(config.Configuration.Colors.Border)

	t.lastSelectRow, _ = t.table.GetSelection()
}

func (t *Tags) GetFocusable() tview.Focusable {
	return t.table.GetFocusable()
}

func (t *Tags) SetData(tags *map[string]int) {
	if tags == nil {
		return
	}
	t.tags = tags
	t.table.Clear()

	i := 0
	for key, count := range *tags {
		t.table.SetCell(i, 0, tableCell(key))
		t.table.SetCell(i, 1, tableCell(fmt.Sprint(count)))
		i += 1
	}

	t.lastSelectRow = -1
}
