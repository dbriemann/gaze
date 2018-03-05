package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

const (
	APIURI = "https://api.thetvdb.com"
)

type pagedData struct {
}

type tvdbClient struct {
}

func tvdbFetchShow(id uint64) (Show, error) {
	s := Show{}
	fmt.Printf("%s> fetching show data from tvdb.com..", pad2)
	resp, body, errs := gorequest.New().Get(APIURI+fmt.Sprintf("/series/%d", id)).Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil {
		fmt.Println(" failed\n")
		return s, errs[0]
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(" failed\n")
		return s, errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
	}

	var incoming struct {
		Data Show `json:"data"`
	}

	err := json.Unmarshal([]byte(body), &incoming)
	if err != nil {
		fmt.Println(" failed\n")
		return s, err
	}

	// TODO fetch episodes
	fmt.Println(" success\n")
	return incoming.Data, nil
}

func (c *tvdbClient) importFavs() error {
	fmt.Printf("%s> importing favorites from thetvdb.com..", pad2)
	resp, body, errs := gorequest.New().Get(APIURI+"/user/favorites").Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil {
		fmt.Println(" failed\n")
		return errs[0]
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(" failed\n")
		return errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
	}

	var incoming struct {
		Data struct {
			Favs []json.Number `json:"favorites"`
		} `json:"data"`
	}

	err := json.Unmarshal([]byte(body), &incoming)
	if err != nil {
		fmt.Println(" failed\n")
		return err
	}

	fmt.Println(" success\n")

	// Add shows to database if they do not exist yet
	newShows := 0
	fmt.Printf("%s> merging favorite shows into database..\n", pad2)
	for _, strid := range incoming.Data.Favs {
		iid, err := strid.Int64()
		if err != nil {
			fmt.Printf("%s> skipping show with ID %s because of error '%s'. ", pad2, strid, err.Error())
			continue
		}
		id := uint64(iid)
		if err := db.addShowByID(id); err != nil {
			fmt.Printf("%s> skipping show with ID %s because of error '%s'. ", pad2, strid, err.Error())
			continue
		}
		newShows++
	}
	fmt.Printf("%s> added %d shows to database\n", pad2, newShows)

	return nil
}

func (c *tvdbClient) ensureLogin() error {
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

func (c *tvdbClient) refresh() error {
	fmt.Printf("%s> refreshing thetvdb.com token..", pad2)
	resp, body, errs := gorequest.New().Get(APIURI+"/refresh_token").Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil || resp.StatusCode != http.StatusOK {
		// If refreshing failed, try logging in.
		fmt.Println(" timed out\n")
		return c.login()
	}

	// Update token and time.
	if err := json.Unmarshal([]byte(body), &configData); err != nil {
		fmt.Println(" failed\n")
		return c.login()
	}

	configData.LastAuth = time.Now()
	configData.save()
	fmt.Println(" success\n")
	return nil
}

func (c *tvdbClient) login() error {
	payload, err := json.Marshal(configData.Auth)
	if err != nil {
		return err
	}
	fmt.Printf("%s> logging in to thetvdb.com..", pad2)
	resp, body, errs := gorequest.New().Post(APIURI + "/login").Send(string(payload)).End()
	if errs != nil {
		fmt.Println(" failed\n")
		return errs[0]
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(" failed\n")
		return errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
	}

	// Update token and time.
	if err := json.Unmarshal([]byte(body), &configData); err != nil {
		fmt.Println(" failed\n")
		return err
	}
	configData.LastAuth = time.Now()
	configData.save()

	fmt.Println(" success\n")
	return nil
}
