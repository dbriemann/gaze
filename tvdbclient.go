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

func tvdbFetchShow(id uint64) (Show, error) {
	s := Show{}
	fmt.Printf("%s> importing show..", pad2)
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
	fmt.Printf(" '%s' (id: %d)\n", incoming.Data.Name, incoming.Data.ID)
	incoming.Data.LastQuery = time.Now().Unix()
	return incoming.Data, nil
}

func tvdbFetchAllEpisodes(s *Show) ([]episode, error) {
	eps := []episode{}

	var incoming struct {
		Links struct {
			First int `json:"first"`
			Last  int `json:"last"`
			//			Next  *int `json:"next"`
			//			Prev  *int `json:"prev"`
		} `json:"links"`
		Data []episode `json:"data"`
	}

	fmt.Printf("%s> importing episodes for show '%s'", pad2, s.Name)
	for page := 1; ; page++ {
		fmt.Print(".")
		resp, body, errs := gorequest.New().Get(APIURI+fmt.Sprintf("/series/%d/episodes?page=%d", s.ID, page)).Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
		if errs != nil {
			fmt.Println(" failed\n")
			return nil, errs[0]
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Println(" failed\n")
			return nil, errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
		}

		err := json.Unmarshal([]byte(body), &incoming)
		if err != nil {
			fmt.Println(" failed\n")
			return nil, err
		}

		eps = append(eps, incoming.Data...)

		if page >= incoming.Links.Last {
			break
		}
	}

	s.LastQuery = time.Now().Unix()
	fmt.Println(" success")

	return eps, nil
}

func tvdbImportFavs() error {
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

	fmt.Println("")

	// Add shows to database if they do not exist yet
	newShows := 0
	fmt.Printf("%s> querying all favorite shows..\n", pad2)
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
	fmt.Printf("%s> added %d shows to database\n\n", pad2, newShows)

	tvdbUpdateAll()
	return nil
}

func tvdbHasShowUpdates(s *Show) (bool, error) {
	if time.Now().Unix() <= s.LastQuery+3600 {
		// Last query for updates is less than an hour ago. Skip to be friendly to servers.
		return false, nil
	}
	//	fmt.Printf("%s> asking if show '%s' (id: %d) has updates..", pad2, s.Name, s.ID)
	resp, _, errs := gorequest.New().Head(APIURI+fmt.Sprintf("/series/%d", s.ID)).Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil {
		//		fmt.Println(" failed\n")
		return false, errs[0]
	}
	if resp.StatusCode != http.StatusOK {
		//		fmt.Println(" failed\n")
		return false, errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
	}

	timeStr := resp.Header.Get("Last-Modified")
	t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", timeStr)
	if err != nil {
		//		fmt.Println(" failed\n")
		return false, err
	}

	hasUpdates := t.Unix() > s.LastUpdated
	s.LastQuery = time.Now().Unix()

	//	fmt.Println(hasUpdates)

	if hasUpdates {
		// Updates are ready to be fetched.
		return true, nil
	}

	// Nothing new..
	return false, nil
}

func tvdbUpdateAll() {

	//1. find all
	//2. update all
	fmt.Printf("%s> querying tvdb.com for updates..", pad2)
	for _, show := range db.Shows {
		up, err := tvdbHasShowUpdates(show)
		// TODO save -> move to db
		if err != nil {
			fmt.Print("x")
			//			fmt.Printf("%s> updating show '%s' (%d) failed.. (error: %s)\n", pad2, show.Name, show.ID, err.Error())
			continue
		} else {
			if up || len(show.EpisodeIDs) == 0 { // Update or first fetch
				// Update we shall.
				err = db.addEpisodesForShow(show)
				if err != nil {
					//				fmt.Printf("%s> updating show '%s' (%d) failed.. (error: %s)\n", pad2, show.Name, show.ID, err.Error())
					continue
				}
			}
			fmt.Print(".")
		}
	}
	fmt.Println(" done\n")
}

func tvdbEnsureLogin() error {
	age := time.Now().Sub(configData.LastAuth)
	if configData.Token != "" && age < time.Hour*24 {
		// An active token exists so we don't need to login.

		if age > 1*time.Hour {
			// Refresh token if older than 1 hour.
			return tvdbRefresh()
		}

		return nil
	}

	// No token exists yet or it has expired -> login.
	return tvdbLogin()
}

func tvdbRefresh() error {
	fmt.Printf("%s> refreshing thetvdb.com token..", pad2)
	resp, body, errs := gorequest.New().Get(APIURI+"/refresh_token").Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil || resp.StatusCode != http.StatusOK {
		// If refreshing failed, try logging in.
		fmt.Println(" timed out\n")
		return tvdbLogin()
	}

	// Update token and time.
	if err := json.Unmarshal([]byte(body), &configData); err != nil {
		fmt.Println(" failed\n")
		return tvdbLogin()
	}

	configData.LastAuth = time.Now()
	configData.save()
	fmt.Println(" success\n")
	return nil
}

func tvdbLogin() error {
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
