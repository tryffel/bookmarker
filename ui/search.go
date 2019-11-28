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
)

type Search struct {
	*tview.InputField
	search func(string)
}

func NewSearch(searchFunc func(query string)) *Search {
	s := &Search{
		InputField: tview.NewInputField(),
		search:     searchFunc}

	s.SetBackgroundColor(config.Configuration.Colors.Background)
	s.SetLabel("Search")
	s.SetLabelWidth(8)
	s.SetLabelColor(config.Configuration.Colors.TextPrimary)
	s.SetDoneFunc(s.Done)
	s.SetBorder(true)
	s.SetBorderColor(config.Configuration.Colors.Border)
	s.SetFieldTextColor(config.Configuration.Colors.BookmarkForm.Text)
	s.SetFieldBackgroundColor(config.Configuration.Colors.BookmarkForm.TextBackground)
	s.SetPlaceholder("my bookmark")
	s.SetPlaceholderTextColor(config.Configuration.Colors.TextPrimaryDim)
	s.SetDoneFunc(s.Done)
	return s
}

func (s *Search) Done(key tcell.Key) {
	if key == tcell.KeyEscape {
		s.SetText("")
		return
	}

	if key != tcell.KeyEnter {
		return
	}

	text := s.GetText()

	s.search(text)
}

func (s *Search) Clear() {
	s.SetText("Search")
}
