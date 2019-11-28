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

package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"sync"
	"tryffel.net/go/bookmarker/config"
	"tryffel.net/go/bookmarker/storage"
	"tryffel.net/go/bookmarker/storage/migrations"
	"tryffel.net/go/bookmarker/storage/models"
	"tryffel.net/go/bookmarker/ui"
)

func main() {
	confFile := flag.String("config", "", "Configuration file location. "+
		"The same directory will be used"+
		"for data also. This can be configured from config file.")

	flag.Parse()

	conf, err := config.ReadConfigFile(*confFile)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	err = config.SaveConfig(conf)
	if err != nil {
		logrus.Error(err)
	}

	level, err := conf.ParseLogLevel()
	if err != nil {
		logrus.Error("Invalid log level format")
		os.Exit(1)
	}

	format := &prefixed.TextFormatter{
		ForceColors:      false,
		DisableColors:    true,
		ForceFormatting:  true,
		DisableTimestamp: false,
		DisableUppercase: false,
		FullTimestamp:    true,
		TimestampFormat:  "",
		DisableSorting:   false,
		QuoteEmptyFields: false,
		QuoteCharacter:   "'",
		SpacePadding:     0,
		Once:             sync.Once{},
	}
	logrus.SetFormatter(format)
	file, err := os.OpenFile(conf.Logfile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0760))
	defer file.Close()
	if err != nil {
		logrus.Error("failed to open log file: ", err.Error())
	}

	logrus.SetOutput(file)

	logrus.Infof("############ %s v%s ############", config.AppName, config.Version)
	logrus.SetLevel(level)

	db, err := storage.NewDatabase(conf.DbFile())
	defer db.Close()
	if err != nil {
		logrus.Errorf("database connection failed: %v", err)
		os.Exit(1)
	}

	// Register user defined metadata
	models.DefaulMetadata = append(models.DefaulMetadata, conf.DefaultMetadata...)
	ui.CustomMetadataFields = conf.DefaultMetadata

	err = migrations.Migrate(db.Engine(), migrations.BookmarkerMigrations)
	if err != nil {
		logrus.Fatal("database migrations failed: %v", err)
		os.Exit(1)
	}

	app := ui.NewWindow(conf.Colors, &conf.Shortcuts, db)
	err = app.Run()
	if err != nil {
		fmt.Printf("Failed to open gui: %v", err)
		os.Exit(1)
	}

}
