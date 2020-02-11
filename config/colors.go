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

package config

import (
	"github.com/gdamore/tcell"
	"tryffel.net/go/twidgets"
)

const (
	colorBackground      = tcell.Color235
	colorModalbackground = tcell.Color239
	colorBackgroundLight = tcell.Color239
	colorText            = tcell.Color23
)

type Colors struct {
	Background               tcell.Color
	TextPrimary              tcell.Color
	TextPrimaryLight         tcell.Color
	TextPrimaryDim           tcell.Color
	SelectionBackground      tcell.Color
	SelectionText            tcell.Color
	Border                   tcell.Color
	BorderFocus              tcell.Color
	ButtonBackground         tcell.Color
	ButtonBackgroundSelected tcell.Color
	ButtonLabel              tcell.Color
	ButtonLabelSelected      tcell.Color
	ModalBackground          tcell.Color
	NavBar                   ColorNavBar
	Bookmarks                ColorBookmarks
	Projects                 ColorProjects
	BookmarkForm             ColorBookmarkForm
	Tags                     ColorTags
	Metadata                 ColorMetadata
	HelpPage                 ColorHelpPage
}

func defaultColors() Colors {
	return Colors{
		Background:               colorBackground,
		TextPrimary:              tcell.Color252,
		TextPrimaryLight:         tcell.Color254,
		TextPrimaryDim:           tcell.Color249,
		SelectionBackground:      tcell.Color23,
		SelectionText:            tcell.Color253,
		Border:                   tcell.Color246,
		BorderFocus:              tcell.Color253,
		ButtonBackground:         tcell.Color241,
		ButtonBackgroundSelected: tcell.Color23,
		ButtonLabel:              tcell.Color254,
		ButtonLabelSelected:      tcell.Color253,
		ModalBackground:          colorModalbackground,
		NavBar:                   defaultColorNavBar(),
		Bookmarks:                defaultColorBookmarks(),
		Projects:                 defaultColorProjects(),
		BookmarkForm:             defaultColorBookmarkform(),
		Tags:                     defaultColorTags(),
		Metadata:                 defaultColorMetadata(),
		HelpPage:                 defaultColorHelpPage(),
	}
}

type ColorNavBar struct {
	Background       tcell.Color
	BackgroundFocus  tcell.Color
	Text             tcell.Color
	TextFocus        tcell.Color
	ButtonBackground tcell.Color
	ButtonFocus      tcell.Color
	Shortcut         tcell.Color
	ShortcutFocus    tcell.Color
}

func defaultColorNavBar() ColorNavBar {
	return ColorNavBar{
		Background:       colorBackground,
		BackgroundFocus:  tcell.Color235,
		Text:             tcell.Color252,
		TextFocus:        tcell.Color253,
		ButtonBackground: colorBackground,
		ButtonFocus:      tcell.Color23,
		Shortcut:         tcell.Color214,
		ShortcutFocus:    tcell.Color214,
	}
}

func (c *ColorNavBar) ToNavBar() *twidgets.NavBarColors {
	return &twidgets.NavBarColors{
		Background:            c.Background,
		BackgroundFocus:       c.BackgroundFocus,
		ButtonBackground:      c.ButtonBackground,
		ButtonBackgroundFocus: c.ButtonFocus,
		Text:                  c.Text,
		TextFocus:             c.TextFocus,
		Shortcut:              c.Shortcut,
		ShortcutFocus:         c.ShortcutFocus,
	}
}

type ColorBookmarks struct {
	Background         tcell.Color
	Background2nd      tcell.Color
	BackgroundSelected tcell.Color
	Text               tcell.Color
	TextSelected       tcell.Color
	HeaderText         tcell.Color
	HeaderBackground   tcell.Color
}

func defaultColorBookmarks() ColorBookmarks {
	return ColorBookmarks{
		Background:         colorBackground,
		Background2nd:      tcell.Color236,
		BackgroundSelected: tcell.Color23,
		Text:               tcell.Color252,
		TextSelected:       tcell.Color253,
		HeaderText:         tcell.Color180,
		HeaderBackground:   tcell.Color235,
	}
}

type ColorProjects struct {
	Background         tcell.Color
	BackgroundSelected tcell.Color
	Text               tcell.Color
	TextSelected       tcell.Color
	Header             tcell.Color
}

func defaultColorProjects() ColorProjects {
	return ColorProjects{
		Background:         colorBackground,
		BackgroundSelected: tcell.Color23,
		Text:               tcell.Color252,
		TextSelected:       tcell.Color253,
		Header:             tcell.Color23,
	}
}

type ColorBookmarkForm struct {
	Background      tcell.Color
	Label           tcell.Color
	Text            tcell.Color
	TextSelected    tcell.Color
	TextBackground  tcell.Color
	TextPlaceHolder tcell.Color
}

func defaultColorBookmarkform() ColorBookmarkForm {
	return ColorBookmarkForm{
		Background:      colorModalbackground,
		Label:           tcell.Color252,
		Text:            tcell.Color187,
		TextSelected:    tcell.Color23,
		TextBackground:  tcell.Color235,
		TextPlaceHolder: tcell.Color247,
	}
}

type ColorTags struct {
	Background         tcell.Color
	BackgroundSelected tcell.Color
	Text               tcell.Color
	TextSelected       tcell.Color
	EmptyTag           tcell.Color
	Count              tcell.Color
}

func defaultColorTags() ColorTags {
	return ColorTags{
		Background:         colorBackground,
		BackgroundSelected: tcell.Color23,
		Text:               tcell.Color228,
		TextSelected:       tcell.Color252,
		EmptyTag:           tcell.Color247,
		Count:              tcell.Color187,
	}

}

type ColorMetadata struct {
	Background         tcell.Color
	BackgroundEditable tcell.Color
	Label              tcell.Color
	Text               tcell.Color
	TextSelected       tcell.Color
	TextBackground     tcell.Color
	TextEdited         tcell.Color
}

func defaultColorMetadata() ColorMetadata {
	return ColorMetadata{
		Background:         tcell.Color236,
		BackgroundEditable: tcell.Color235,
		Label:              tcell.Color252,
		Text:               tcell.Color187,
		TextSelected:       tcell.Color23,
		TextBackground:     tcell.Color238,
		TextEdited:         tcell.Color187,
	}
}

type ColorHelpPage struct {
	Background tcell.Color
	Text       tcell.Color
	Headers    tcell.Color
}

func defaultColorHelpPage() ColorHelpPage {
	return ColorHelpPage{
		Background: colorModalbackground,
		Text:       tcell.Color252,
		Headers:    tcell.Color228,
	}
}
