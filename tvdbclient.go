package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/parnurzeal/gorequest"
)

const (
	APIURI = "https://api.thetvdb.com"
)

type TVDBClient struct {
	// TODO add database
}

func (c *TVDBClient) ensureLogin() error {
	age := time.Now().Sub(configData.LastAuth)
	if configData.Token != "" && age < time.Hour*24 {
		// An active token exists so we don't need to login.

		if age > 1*time.Hour {
			// Refresh token if older than 1 hour.
			return c.refresh()
		}

		return nil
	}

	// No token exists yet or it has expired -> login.
	return c.login()
}

func (c *TVDBClient) refresh() error {
	_, _, errs := gorequest.New().Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).Get(APIURI + "/refresh_token").End()
	if errs != nil {
		return errs[0]
	}
	configData.LastAuth = time.Now()
	configData.save()
	return nil
}

func (c *TVDBClient) login() error {
	payload, err := json.Marshal(configData.Auth)
	if err != nil {
		return err
	}
	fmt.Printf("%s> logging in to thetvdb.com..", pad2)
	_, body, errs := gorequest.New().Post(APIURI + "/login").Send(string(payload)).End()
	if errs != nil {
		fmt.Println(" failed")
		return errs[0]
	}

	// Update token and time.
	if err := json.Unmarshal([]byte(body), &configData); err != nil {
		fmt.Println(" failed")
		return err
	}
	configData.LastAuth = time.Now()
	configData.save()

	fmt.Println(" success")
	return nil
}
