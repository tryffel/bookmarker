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
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"runtime"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/storage"
)

const (
	helpText = `
Bookmarker help page. Press <esc> to close this window.
Navigation
* J/K (row up/down)
* Tab (move between widgets / metadata fields)

`

	logo = `
______             _                         _             
| ___ \           | |                       | |            
| |_/ / ___   ___ | | ___ __ ___   __ _ _ __| | _____ _ __ 
| ___ \/ _ \ / _ \| |/ / '_ ' _ \ / _' | '__| |/ / _ \ '__|
| |_/ / (_) | (_) |   <| | | | | | (_| | |  |   <  __/ |   
\____/ \___/ \___/|_|\_\_| |_| |_|\__,_|_|  |_|\_\___|_|   
`
)

type Help struct {
	*tview.TextView
	doneFunc func()
	visible  bool
}

func (h *Help) SetDoneFunc(doneFunc func()) {
	h.doneFunc = doneFunc
}

func (h *Help) SetVisible(visible bool) {
	h.visible = visible
}

func (h *Help) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			h.doneFunc()
		} else {
			h.TextView.InputHandler()(event, setFocus)
		}
	}
}

func NewHelp() *Help {
	h := &Help{
		TextView: tview.NewTextView(),
		doneFunc: nil,
		visible:  false,
	}

	colors := config.Configuration.Colors
	h.TextView.SetBorder(true)
	h.TextView.SetBorderColor(colors.BorderFocus)
	h.TextView.SetBackgroundColor(colors.Background)
	h.TextView.SetTextColor(colors.TextPrimary)
	h.TextView.SetBorderPadding(1, 1, 1, 1)

	text := fmt.Sprintf("%s\nv %s\n%s", logo, config.Version, helpText)
	h.TextView.SetText(text)

	return h
}

func (h *Help) Update(stats *storage.Statistics) {
	text := fmt.Sprintf("%s\nv %s\n%s", logo, config.Version, helpText)

	runStats := runtime.MemStats{}
	runtime.ReadMemStats(&runStats)

	timeFormat := "2006-01-02 15:04:05"
	text += fmt.Sprintf("Statistics:\nBookmarks: %d\nARchived: %d\nTags: %d\nProjects: %d\nLast Bookmark: %s\n",
		stats.Bookmarks, stats.Archived, stats.Tags, stats.Projects, stats.LastBookmark.Format(timeFormat))

	text += fmt.Sprintf("Memory: %s\n", formatBytes(runStats.Alloc))
	text += fmt.Sprintf("Config location: %s\n", config.Configuration.ConfigDir())
	h.SetText(text)
}

func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprint(bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%d KiB", bytes/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%d MiB", bytes/1024/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%d GiB", bytes/1024/1024/1024)
	}
	return ""
}
