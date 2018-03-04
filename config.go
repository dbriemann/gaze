package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	configFile = "config.json"
)

type config struct {
	Auth     auth      `json:"auth"`
	Token    string    `json:"token"`
	LastAuth time.Time `json:"lastauth"`
}

type auth struct {
	ApiKey   string `json:"apikey"`
	UserKey  string `json:"userkey"`
	UserName string `json:"username"`
}

var (
	workDir    = ""
	configData config
)

func (c *config) save() {
	b, err := json.MarshalIndent(*c, "", "\t")
	if err != nil {
		bye(fmt.Sprintf(pad2+"> could not create auth file: %s\n", err.Error()), 1)
	}

	path := filepath.Join(workDir, configFile)

	err = ioutil.WriteFile(path, b, 0600)
	if err != nil {
		bye(fmt.Sprintf(pad2+"> could not save auth file: %s\n", err.Error()), 1)
	}
}

func (c *config) hasAuthData() bool {
	if c.Auth.ApiKey == "" ||
		c.Auth.UserKey == "" ||
		c.Auth.UserName == "" {
		return false
	}

	return true
}

func mountDataDir(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	finfo, err := os.Stat(abs)

	// If the path does not exist.
	if err != nil && os.IsNotExist(err) {
		// Create directory.
		err = os.MkdirAll(abs, 0755)
		if err != nil {
			return err
		}
		fmt.Println(pad2+"> Created working directory at", abs)
		workDir = abs
	} else {
		// Test if it is a directory.
		if !finfo.IsDir() {
			return errors.New(abs + " expected directory but got file")
		}
		workDir = abs
	}

	return err
}

func loadConfig() {
	fpath := filepath.Join(workDir, configFile)
	// Check if config file exists and if not create it.
	finfo, err := os.Stat(fpath)

	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file.
			configData = config{}
			configData.save()
			return
		} else {
			bye(fmt.Sprintf(pad2+"> something bad happened (%s)\n", err.Error()), 1)
		}
	}

	if !finfo.IsDir() {
		// Read config file.
		raw, err := ioutil.ReadFile(fpath)
		if err != nil {
			bye(fmt.Sprintf(pad2+"> could not read file %s with error: %s", fpath, err.Error()), 1)
		}

		err = json.Unmarshal(raw, &configData)
		if err != nil {
			bye(fmt.Sprintf(pad2+"> could not read file %s with error: %s", fpath, err.Error()), 1)
		}
	} else {
		bye(pad2+"> %s is a directory but should be a file", 1)
	}
}
