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
	"fmt"
	"github.com/rivo/tview"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/storage/models"
)

type DeleteBookmark struct {
	*tview.Modal
	done func(bool)
}

func (d *DeleteBookmark) SetDoneFunc(doneFunc func()) {
}

func (d *DeleteBookmark) SetVisible(visible bool) {
}

func NewDeleteBookmark(doneFunc func(bool), bookmark *models.Bookmark) *DeleteBookmark {
	d := &DeleteBookmark{
		Modal: tview.NewModal(),
		done:  doneFunc,
	}

	colors := config.Configuration.Colors.BookmarkForm
	col := config.Configuration.Colors
	d.Modal.SetBackgroundColor(colors.Background)
	d.Modal.SetBorder(true)
	d.Modal.SetBorderColor(col.BorderFocus)
	d.Modal.SetTextColor(colors.Text)
	d.Modal.SetTitle("Delete Bookmark")

	d.SetText(fmt.Sprintf("Are you sure you want to delete bookmark \"%s\"", bookmark.Name))
	d.Modal.SetButtonBackgroundColor(col.ButtonBackground)
	d.Modal.SetButtonTextColor(col.ButtonLabel)

	d.AddButtons([]string{"Delete", "Cancel"})
	d.Modal.SetDoneFunc(d.doneFunc)
	return d
}

func (d *DeleteBookmark) doneFunc(index int, label string) {
	if d.done == nil {
		return
	}
	if label == "Delete" {
		d.done(true)
	} else {
		d.done(false)
	}
}
