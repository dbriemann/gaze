package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kr/pretty"
	"github.com/parnurzeal/gorequest"
)

const (
	APIURI = "https://api.thetvdb.com"
)

type db struct {
	Shows       map[uint64]show `json:"shows"`
	LastUpdated uint64          `json:"lastUpdated"`
}

type show struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"seriesName"`
	Status      string    `json:"status"`
	Network     string    `json:"network"`
	FirstAired  string    `json:"firstAired"`
	Runtime     uint32    `json:"runtime,string"`
	Overview    string    `json:"overview"`
	LastUpdated uint64    `json:"lastUpdated"`
	AirDay      string    `json:"airsDayOfWeek"`
	AirTime     string    `json:"airsTime"`
	Rating      float32   `json:"siteRating"`
	RatingCount uint32    `json:"siteRatingCount"`
	Episodes    []episode `json:"episodes"`
}

type episode struct {
	ID          uint64 `json:"id"`
	Name        string `json:"episodeName"`
	Number      uint32 `json:"airedEpisodeNumber"`
	Season      uint32 `json:"airedSeason"`
	FirstAired  string `json:"firstAired"`
	LastUpdated uint64 `json:"lastUpdated"`
	Overview    string `json:"overview"`

	Watched bool `json:"watched"`
}

type tvdbClient struct {
	data db
}

func (c *tvdbClient) importFavs() error {
	fmt.Printf("%s> importing favorites from thetvdb.com..", pad2)
	_, body, errs := gorequest.New().Get(APIURI+"/user/favorites").Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil {
		fmt.Println(" failed\n")
		return errs[0]
	}

	pretty.Println(body)

	var d struct {
		favs []show `json:"favorites"`
	}

	err := json.Unmarshal([]byte(body), &d)
	if err != nil {
		fmt.Println(" failed\n")
		return err
	}

	fmt.Println(" success\n")

	pretty.Println(d.favs)

	// TODO - continue here

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
	_, body, errs := gorequest.New().Post(APIURI + "/login").Send(string(payload)).End()
	if errs != nil {
		fmt.Println(" failed\n")
		return errs[0]
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
