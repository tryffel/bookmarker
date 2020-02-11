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
	"tryffel.net/go/bookmarker/storage"
)

//Modify is a modal that operates on bulk of bookmarks defined with filter and modifiers
type Modify struct {
	*tview.Form
	doneFunc   func()
	modifyFunc func(filter *storage.Filter, modifier *storage.Modifier) (int, error)

	filter *tview.InputField
	key    *tview.InputField
	value  *tview.InputField
	status *tview.InputField
}

func (m *Modify) SetDoneFunc(doneFunc func()) {
}

func (m *Modify) SetVisible(visible bool) {
}

func NewModify(modifyFunc func(filter *storage.Filter, modifier *storage.Modifier) (int, error)) *Modify {
	m := &Modify{
		Form:       tview.NewForm(),
		doneFunc:   nil,
		modifyFunc: modifyFunc,
		filter:     tview.NewInputField().SetLabel("Filter").SetPlaceholder("project:bookmarks"),
		key:        tview.NewInputField().SetLabel("Key").SetPlaceholder("archived"),
		value:      tview.NewInputField().SetLabel("Value").SetPlaceholder("true"),
		status: tview.NewInputField().SetLabel("Status").SetAcceptanceFunc(func(string, rune) bool {
			return false
		}),
	}
	colors := config.Configuration.Colors.BookmarkForm

	m.SetTitle("Bulk modify")
	m.SetTitleColor(colors.Text)

	m.SetBorder(true)
	m.SetBorderColor(config.Configuration.Colors.Border)
	m.SetBackgroundColor(colors.Background)
	m.SetLabelColor(colors.Label)
	m.SetFieldBackgroundColor(colors.TextBackground)
	m.SetFieldTextColor(colors.Text)

	warning := tview.NewInputField().SetLabel("[::u]Warning[::-]").
		SetText("This is experimental feature. Use at your own risk (backup database file first)")
	//disable edits
	warning.SetAcceptanceFunc(func(string, rune) bool { return false })
	warning.SetBackgroundColor(config.Configuration.Colors.ModalBackground)
	m.AddFormItem(warning)

	m.AddFormItem(m.filter)
	m.AddFormItem(m.key)
	m.AddFormItem(m.value)
	m.AddFormItem(m.status)

	m.AddButton("Execute", m.save)
	return m
}

func (m *Modify) save() {
	m.status.SetText("")
	filter, err := storage.NewFilter(m.filter.GetText())
	if err != nil {
		m.status.SetText(fmt.Errorf("Error: invalid filter: %v", err).Error())
		return
	}

	modifier, err := storage.NewModifier(m.key.GetText(), m.value.GetText())
	if err != nil {
		m.status.SetText(fmt.Errorf("Error: invalid modifier: %v", err).Error())
		return
	}

	if m.modifyFunc != nil {
		count, error := m.modifyFunc(filter, modifier)
		if error == nil {
			m.status.SetText(fmt.Sprintf("%d Bookmarks modified", count))
		} else {
			m.status.SetText(fmt.Errorf("Error: %v", err).Error())
		}
	}
}
