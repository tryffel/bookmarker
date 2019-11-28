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
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path"
)

var configFile = AppNameLower + ".toml"
var logFile = AppNameLower + ".log"
var dbFile = AppNameLower + ".sqlite"

//ReadConfigFile reads config file from given file. If file is empty, use default location provided by os
func ReadConfigFile(file string) (*ApplicationConfig, error) {
	if file == "" {
		return readDefaultConfig()
	}

	conf := defaultConfig()
	err := EnsureConfigDirExists()
	if err != nil {
		return nil, err
	}
	Configuration = conf

	// Error caught in previous call
	//dir, _ := GetConfigDirectory()
	//dir = path.Join(dir, AppNameLower)

	dir := path.Dir(file)

	conf.configDir = dir
	conf.configFile = file
	conf.Log = path.Join(dir, logFile)
	conf.DataBase = path.Join(dir, dbFile)

	err = EnsureFileExists(conf.configFile)
	if err != nil {
		return nil, err
	}
	err = EnsureFileExists(conf.Log)
	if err != nil {
		return nil, err
	}
	err = EnsureFileExists(conf.DataBase)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(conf.configFile)

	_, err = toml.DecodeReader(f, &conf)
	if err != nil {
		return conf, fmt.Errorf("failed to read config file: %v", err)
	}

	conf.Log = path.Join(dir, logFile)
	conf.DataBase = path.Join(dir, dbFile)

	return conf, nil
}

func readDefaultConfig() (*ApplicationConfig, error) {
	conf := defaultConfig()
	err := EnsureConfigDirExists()
	if err != nil {
		return nil, err
	}

	Configuration = conf

	// Error caught in previous call
	dir, _ := GetConfigDirectory()
	dir = path.Join(dir, AppNameLower)

	conf.configDir = dir
	conf.configFile = path.Join(dir, configFile)
	conf.Log = path.Join(dir, logFile)
	conf.DataBase = path.Join(dir, dbFile)

	err = EnsureFileExists(conf.configFile)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(conf.configFile)

	_, err = toml.DecodeReader(file, &conf)
	if err != nil {
		return conf, fmt.Errorf("failed to read config file: %v", err)
	}

	err = EnsureFileExists(conf.Log)
	if err != nil {
		return nil, err
	}
	err = EnsureFileExists(conf.DataBase)
	if err != nil {
		return nil, err
	}

	//conf.Log = path.Join(dir, logFile)
	//conf.DataBase = path.Join(dir, dbFile)

	return conf, nil

}

func SaveConfig(conf *ApplicationConfig) error {
	file, err := os.OpenFile(conf.configFile, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	return toml.NewEncoder(file).Encode(conf)
}

func GetConfigDirectory() (string, error) {
	return os.UserConfigDir()
}

func EnsureConfigDirExists() error {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := path.Join(userConfig, AppNameLower)
	dirExists, err := DirectoryExists(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if dirExists {
		return nil
	} else {
		err = CreateDirectory(dir)
		return err
	}
}

func EnsureFileExists(name string) error {
	exists, err := FileExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return CreateFile(name)
}

func DirectoryExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return true, nil
	}
	return false, fmt.Errorf("not directory")
}

func CreateDirectory(dir string) error {
	return os.Mkdir(dir, 0760)
}

func CreateFile(name string) error {
	file, err := os.Create(name)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
