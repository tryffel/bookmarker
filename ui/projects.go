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

type Projects struct {
	table   *tview.Table
	project []*models.Project

	rows []*models.Project

	selected   bool
	selectFunc func(bookmark *models.Project)
}

func NewProjects() *Projects {
	p := &Projects{table: tview.NewTable()}
	p.rows = []*models.Project{}
	colors := config.Configuration.Colors.Projects
	p.table.SetTitle("Projects")
	p.table.SetTitleColor(colors.Text)
	p.table.SetBorder(true)
	p.table.SetBorderColor(config.Configuration.Colors.Border)
	p.table.SetBackgroundColor(colors.Background)
	p.table.SetBorderColor(config.Configuration.Colors.Border)
	p.table.SetSelectedFunc(p.selectProject)
	p.table.SetSelectable(true, false)
	p.table.SetSelectedStyle(config.Configuration.Colors.Projects.Text,
		config.Configuration.Colors.Projects.Background, 0)

	return p
}

func (p *Projects) SetSelectFunc(selectFunc func(project *models.Project)) {
	p.selectFunc = selectFunc
}

func (p *Projects) Draw(screen tcell.Screen) {
	p.table.Draw(screen)
}

func (p *Projects) GetRect() (int, int, int, int) {
	return p.table.GetRect()
}

func (p *Projects) SetRect(x, y, width, height int) {
	p.table.SetRect(x, y, width, height)
}

func (p *Projects) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.table.InputHandler()
}

func (p *Projects) Focus(delegate func(p tview.Primitive)) {
	p.table.Focus(delegate)
	p.table.SetBorderColor(config.Configuration.Colors.BorderFocus)
	p.table.SetSelectedStyle(config.Configuration.Colors.Bookmarks.TextSelected,
		config.Configuration.Colors.Bookmarks.BackgroundSelected, 0)
}

func (p *Projects) Blur() {
	p.table.Blur()
	p.table.SetBorderColor(config.Configuration.Colors.Border)
	if !p.selected {
		p.table.SetSelectedStyle(config.Configuration.Colors.Projects.Text,
			config.Configuration.Colors.Projects.Background, 0)
	}
}

func (p *Projects) GetFocusable() tview.Focusable {
	return p.table.GetFocusable()
}

func (p *Projects) SetData(projects []*models.Project) {
	p.project = projects
	p.rows = []*models.Project{}
	p.table.Clear()
	p.table.SetCell(0, 0, tableHeaderCell("Name"))
	p.table.SetCell(0, 1, tableHeaderCell("Count"))
	p.table.SetCell(1, 0, tableCell("-"))
	p.table.SetCell(1, 1, tableCell(""))
	p.table.SetOffset(1, 0)
	p.table.SetFixed(1, 0)

	totalCount := 0

	indent := ""
	index := 2

	for _, project := range projects {
		totalCount += project.TotalCount()
		index = p.addProject(project, index, indent)
		index += 1
	}
	p.table.SetCell(1, 1, tableCell(fmt.Sprint(totalCount)))
	p.table.Select(1, 0)
}

func (p *Projects) addProject(project *models.Project, index int, indent string) int {
	p.table.SetCell(index, 0, tableCell(indent+project.Name))
	p.table.SetCell(index, 1, tableCell(fmt.Sprint(project.TotalCount())))
	p.rows = append(p.rows, project)

	for _, child := range project.Children {
		index = p.addProject(child, index+1, indent+"  ")
	}
	return index
}

func (p *Projects) selectProject(row, col int) {
	p.selected = true

	if p.selectFunc != nil {
		if row == 1 {
			p.selectFunc(nil)
		} else {
			project := p.rows[row-2]
			p.selectFunc(project)
		}
	}
}
