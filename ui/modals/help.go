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
	"strings"
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

	logo = "" +
		`______             _                         _             
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

	page       int
	totalPages int
	stats      *storage.Statistics
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
		} else if key == tcell.KeyLeft {
			if h.page > 0 {
				h.page -= 1
				h.setContent()
			}
		} else if key == tcell.KeyRight {
			if h.page < h.totalPages-1 {
				h.page += 1
				h.setContent()
			}
		} else {
			h.TextView.InputHandler()(event, setFocus)
		}
	}
}

func (h *Help) setContent() {
	title := ""
	got := ""
	switch h.page {
	case 0:
		got = h.mainPage()
		title = "About"
	case 1:
		got = h.shortcutsPage()
		title = "Usage"
	case 2:
		got = h.searchPage()
		title = "Searching"
	case 3:
		got = h.statsPage()
		title = "Info"
	default:
	}

	if title != "" {
		title = "[yellow::b]" + title + "[-::-]"
	}

	if got != "" {
		h.Clear()
		text := fmt.Sprintf("< %d / %d > %s \n\n", h.page+1, h.totalPages, title)
		text += got
		h.SetText(text)
		h.ScrollToBeginning()
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
	h.TextView.SetBackgroundColor(colors.HelpPage.Background)
	h.TextView.SetTextColor(colors.HelpPage.Text)
	h.TextView.SetBorderPadding(1, 1, 2, 2)
	h.TextView.SetDynamicColors(true)
	h.TextView.SetWordWrap(true)

	h.totalPages = 4
	h.setContent()
	return h
}

func (h *Help) statsPage() string {
	var stats *storage.Statistics
	if h.stats != nil {
		stats = h.stats
	} else {
		stats = &storage.Statistics{}
	}

	text := "[yellow]Statistics[-]\n"

	runStats := runtime.MemStats{}
	runtime.ReadMemStats(&runStats)

	timeFormat := "2006-01-02 15:04:05"
	text += fmt.Sprintf("Bookmarks: %d\nArchived: %d\nTags: %d\nProjects: %d\nLast Bookmark: %s\n",
		stats.Bookmarks, stats.Archived, stats.Tags, stats.Projects, stats.LastBookmark.Format(timeFormat))

	text += fmt.Sprintf("Memory: %s\n", formatBytes(runStats.Alloc))

	text += "\n[yellow]Additional info[-]\n"
	text += fmt.Sprintf("Data location: %s\n", config.Configuration.ConfigDir())
	text += fmt.Sprintf("Full text search engine supported: %v\n", stats.FullTextSearchSupported)

	text += fmt.Sprintf("Metadata keys: \n  * %s", strings.Join(h.stats.MetadataKeys, "\n  * "))
	return text
}

func (h *Help) Update(stats *storage.Statistics) {
	h.stats = stats
}

func (h *Help) mainPage() string {
	text := fmt.Sprintf("%s\n[yellow]v%s[-]\n\n", logo, config.Version)
	text += "License: Apache-2.0, http://www.apache.org/licenses/LICENSE-2.0"
	return text
}

func (h *Help) searchPage() string {
	return "" +
		`[yellow]Full text search[-]
If Bookmarker was built with full text search support, 
plain search queries (no filters) will result in full text queries. Results are then highlighted.
Full text queries can be any text:
'[#00d7ff]my awesome site[-]'
this will return any bookmark that contains phrase containing each word in query.
You can use AND/OR/NOT clauses to modify query:
'[#00d7ff](mypage AND com) OR mypage.com))[-]'
Wildcards are supported:
'[#00d7ff]mypag*[-]'
Exact matches:
'[#00d7ff]"mypage that contains a"[-]'

[yellow]Filters[-]
Any field can be filtered with key:value format:
'[#00d7ff]author:jack[-]'
Multiple filters can be applied:
'[#00d7ff]author:jack link:mypage.com[-]'
Negations can be applied with preceding '-':
'[#00d7ff]author:jack -link:mypage.com[-]'

Matching field exactly can be done by enclosing value with '. e.g.:
'[#00d7ff]link:'mypage.com'[-]'
Would match exactly link mypage.com
`
}

func (h *Help) shortcutsPage() string {
	return `[yellow]Movement[-]:
* Up/Down: Key up / down
* VIM-like keys: 
	* Up / Down: J / K 
	* Top / Bottom: g / G 
	* Page Up / Down: Ctrl+F / Ctrl+B
* Switch panels: Tab

[yellow]Forms[-]:
* Tab / Shift-Tab moves between form fields

[yellow]Search[-]:
* Open search panel: Ctrl-D
* Search: Enter
* Cancel: Escape

[yellow]Metadata[-]:
* Ctrl-space opens metadata viewer for selected bookmark

[yellow]Sorting[-]:
* Navigate to any column header and press enter to sort either ascending or descending
`
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
