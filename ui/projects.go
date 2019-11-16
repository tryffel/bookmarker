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
	tree    *tview.TreeView
	project []*models.Project
}

func NewProjects() *Projects {
	p := &Projects{tree: tview.NewTreeView()}
	colors := config.Configuration.Colors.Projects
	p.tree.SetTitle("Projects")
	p.tree.SetTitleColor(colors.Text)
	p.tree.SetBorder(true)
	p.tree.SetBorderColor(config.Configuration.Colors.Border)
	p.tree.SetBackgroundColor(colors.Background)
	p.tree.SetBorderColor(config.Configuration.Colors.Border)
	return p
}

func (p *Projects) Draw(screen tcell.Screen) {
	p.tree.Draw(screen)
}

func (p *Projects) GetRect() (int, int, int, int) {
	return p.tree.GetRect()
}

func (p *Projects) SetRect(x, y, width, height int) {
	p.tree.SetRect(x, y, width, height)
}

func (p *Projects) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.tree.InputHandler()
}

func (p *Projects) Focus(delegate func(p tview.Primitive)) {
	p.tree.Focus(delegate)
	p.tree.SetBorderColor(config.Configuration.Colors.BorderFocus)
}

func (p *Projects) Blur() {
	p.tree.Blur()
	p.tree.SetBorderColor(config.Configuration.Colors.Border)
}

func (p *Projects) GetFocusable() tview.Focusable {
	return p.tree.GetFocusable()
}

func (p *Projects) SetData(projects []*models.Project) {
	p.project = projects

	top := newTreeNode("Projects")

	// Root nodes
	for _, root := range projects {
		rootItem := newTreeNode(fmt.Sprintf("%s %d", root.Name, root.TotalCount()))
		addTree(root, rootItem)
		top.AddChild(rootItem)
	}

	p.tree.SetRoot(top)
}

// Add tree adds recursive tree to node
func addTree(project *models.Project, node *tview.TreeNode) {
	node.SetText(fmt.Sprintf("%s %d", project.Name, project.TotalCount()))
	node.SetColor(config.Configuration.Colors.Projects.Text)
	if project.Children == nil {
		return
	}

	for _, v := range project.Children {
		child := tview.NewTreeNode(v.Name)
		addTree(v, child)
		node.AddChild(child)
	}
}

func newTreeNode(text string) *tview.TreeNode {
	node := tview.NewTreeNode(text)
	node.SetColor(config.Configuration.Colors.Bookmarks.Text)
	return node
}
