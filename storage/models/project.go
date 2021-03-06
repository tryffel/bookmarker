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

package models

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

const TreeIndent = "   "

type Project struct {
	Name     string
	Children []*Project
	Parent   *Project
	Count    int
}

func NewProject(name string) *Project {
	p := &Project{
		Name:     name,
		Children: []*Project{},
		Parent:   nil,
	}
	return p
}

//String returns string presentation of project.
//Only print project and its parents:
// c.string -> a.b.c, a.string -> a
func (p *Project) String() string {
	return p.FullName()
}

func (p *Project) FullName() string {
	if p.Parent == nil {
		return p.Name
	}
	return p.Parent.FullName() + "." + p.Name
}

func (p *Project) sortChildren(recurse bool) {
	if len(p.Children) == 0 {
		return
	}
	sort.SliceStable(p.Children, func(i, j int) bool {
		return p.Children[i].Name < p.Children[j].Name
	})

	if recurse {
		for _, v := range p.Children {
			v.sortChildren(true)
		}
	}
}

func (p *Project) TotalCount() int {
	count := p.Count

	for _, v := range p.Children {
		count += v.TotalCount()
	}
	return count
}

// ParseTrees parses array if strings and array of counts into tree of projects
// Data: e.g. ["project.a", "project.b", "project.a.b.c"]
// Count: e.g. [10,10,10]
func ParseTrees(data []string, counts []int) []*Project {
	if len(data) != len(counts) {
		logrus.Warningf("parse trees data and counts length does not match, ignore counts")
		counts = make([]int, len(data))
	}

	root := NewProject("root")
	for i, v := range data {
		text := strings.Split(v, ".")
		text = append([]string{"root"}, text...)
		ok := root.parseSingle(text, counts[i])
		if !ok {
			fmt.Printf("%s failed", v)
		}
	}
	for _, v := range root.Children {
		v.Parent = nil
	}
	return root.Children
}

func (p *Project) parseSingle(nodes []string, count int) bool {
	if len(nodes) == 0 {
		return false
	}
	if len(nodes) == 1 {
		if nodes[0] == p.Name {
			p.Count = count
			return true
		}
		return false
	}

	exists := false

	// Try for existing children
	for _, v := range p.Children {
		if nodes[1] == v.Name {
			if v.parseSingle(nodes[1:], count) {
				exists = true
				break
			}
		}
	}
	// Create new child
	if !exists {
		child := NewProject(nodes[1])
		child.Parent = p
		slice := make([]string, 0)
		if len(nodes) > 2 {
			slice = nodes[1:]
		} else {
			slice = []string{nodes[1]}
		}
		ok := child.parseSingle(slice, count)
		p.Children = append(p.Children, child)
		if ok {
			return true
		}
		return false
	}
	return true
}

func (p *Project) PrintChildren() string {
	text := strings.Join(p.printChildren(0), "\n")
	return text
}

func (p *Project) printChildren(indent int) []string {
	text := p.Name
	ind := ""
	for i := 0; i < indent; i++ {
		ind += TreeIndent
	}

	if indent > 0 {
		text = ind + text
	}

	if len(p.Children) == 0 {
		return []string{text}
	}

	children := []string{}
	for _, v := range p.Children {
		children = append(children, v.printChildren(indent+1)...)
	}

	return append([]string{text}, children...)
}
