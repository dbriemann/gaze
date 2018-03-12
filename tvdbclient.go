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
	APIURI                  = "https://api.thetvdb.com"
	secondsUntilUpdateCheck = 60 * 60 * 4 // No update checks before 4 hours have passed since last check.
)

func tvdbFetchShow(id uint64, isUpdate bool) (Show, error) {
	s := Show{}
	if !isUpdate {
		fmt.Printf("%s> importing show..", pad2)
	}

	resp, body, errs := gorequest.New().Get(APIURI+fmt.Sprintf("/series/%d", id)).Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil {
		if !isUpdate {
			fmt.Println(" failed\n")
		}
		return s, errs[0]
	}
	if resp.StatusCode != http.StatusOK {
		if !isUpdate {
			fmt.Println(" failed\n")
		}
		return s, errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
	}

	var incoming struct {
		Data Show `json:"data"`
	}

	err := json.Unmarshal([]byte(body), &incoming)
	if err != nil {
		if !isUpdate {
			fmt.Println(" failed\n")
		}
		return s, err
	}

	if !isUpdate {
		fmt.Printf(" '%s' (id: %d)\n", incoming.Data.Name, incoming.Data.ID)
	}
	//	incoming.Data.LastQuery = time.Now().Unix()
	return incoming.Data, nil
}

func tvdbFetchAllEpisodes(s *Show) ([]*episode, error) {
	eps := []*episode{}

	var incoming struct {
		Links struct {
			First int `json:"first"`
			Last  int `json:"last"`
		} `json:"links"`
		Data []*episode `json:"data"`
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

	//	s.LastQuery = time.Now().Unix()
	fmt.Println(" success")

	return eps, nil
}

func tvdbFetchFavorites() ([]uint64, error) {
	ids := []uint64{}

	fmt.Printf("%s> fetching favorites from thetvdb.com..", pad2)
	resp, body, errs := gorequest.New().Get(APIURI+"/user/favorites").Set("Authorization", fmt.Sprintf("Bearer %s", configData.Token)).End()
	if errs != nil {
		fmt.Println(" failed\n")
		return ids, errs[0]
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(" failed\n")
		return ids, errors.New("reply from thetvdb.com has wrong status code " + strconv.Itoa(resp.StatusCode))
	}

	var incoming struct {
		Data struct {
			Favs []json.Number `json:"favorites"`
		} `json:"data"`
	}

	err := json.Unmarshal([]byte(body), &incoming)
	if err != nil {
		fmt.Println(" failed\n")
		return ids, err
	}

	fmt.Println("")

	for _, strid := range incoming.Data.Favs {
		id, err := strid.Int64()
		if err != nil {
			fmt.Printf("%s> skipping show with ID %s because of error '%s'. ", pad2, strid, err.Error())
			continue
		}
		ids = append(ids, uint64(id))
	}

	return ids, nil

}

func tvdbHasShowUpdates(s *Show) (bool, error) {
	if time.Now().Unix() <= s.LastQuery+secondsUntilUpdateCheck {
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
