/*
 *
 *  Copyright 2019 Tero Vierimaa
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package config

import (
	"github.com/sirupsen/logrus"
)

type ApplicationConfig struct {
	LogLevel               string   `toml:"log_level"`
	HideArchived           bool     `toml:"default_hide_archived"`
	DefaultMetadata        []string `toml:"default_metadata_fields"`
	DataBase               string   `toml:"database_file"`
	Log                    string   `toml:"log_file"`
	AutoComplete           bool     `toml:"autocomplete"`
	AutoCompleteMaxResults int      `toml:"autocomplete_max_results"`
	EnableFullTextSearch   bool     `toml:"full_text_search"`
	Colors                 Colors
	Shortcuts              Shortcuts
	configDir              string
	configFile             string
}

var Configuration *ApplicationConfig = &ApplicationConfig{}

func (a *ApplicationConfig) ParseLogLevel() (logrus.Level, error) {
	return logrus.ParseLevel(a.LogLevel)
}

func (a *ApplicationConfig) DbFile() string {
	return a.DataBase
}

func (a *ApplicationConfig) Logfile() string {
	return a.Log
}

func (a *ApplicationConfig) ConfigDir() string {
	return a.configDir
}

//Default configuration which config file overwrites
func defaultConfig() *ApplicationConfig {
	conf := &ApplicationConfig{
		LogLevel:               "debug",
		HideArchived:           true,
		DefaultMetadata:        []string{"Author", "Published At", "Language", "Ipfs", "Class", "Title"},
		AutoComplete:           true,
		AutoCompleteMaxResults: 20,
		EnableFullTextSearch:   true,
		Colors:                 defaultColors(),
		Shortcuts:              defaultShortcuts(),
	}
	return conf
}
