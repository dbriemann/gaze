package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	db *database
)

type database struct {
	Shows       map[uint64]*Show    `json:"shows"`
	Episodes    map[uint64]*episode `json:"episodes"`
	LastUpdated uint64              `json:"lastUpdated"`
	Version     uint32              `json:"version"`
}

func OpenDB() *database {
	db := &database{
		Shows:       map[uint64]*Show{},
		Episodes:    map[uint64]*episode{},
		LastUpdated: uint64(time.Now().Unix()),
	}
	fpath := filepath.Join(workDir, dbFile)

	// Test if database file exists.
	finfo, err := os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			// If not create a dummy.
			db.save()
			fmt.Println(pad2+"> created database file at", fpath)
			fmt.Println()
		} else {
			// Strange error -> exit.
			bye(fmt.Sprintf("could not open database: %s", err.Error()), 1)
		}
	} else {
		if finfo.IsDir() {
			// Path is a directory but should be a file -> exit.
			bye(fmt.Sprintf("could not open database: %s", ErrDirNotFile.Error()), 1)
		} else {
			// Read database file
			raw, err := ioutil.ReadFile(fpath)
			if err != nil {
				bye(fmt.Sprintf("could not open database: %s", err.Error()), 1)
			}

			err = json.Unmarshal(raw, db)
			if err != nil {
				bye(fmt.Sprintf("could not open database: %s", err.Error()), 1)
			}
		}
	}

	// Here we have a sane database object.
	return db
}

func (db *database) save() {
	b, err := json.MarshalIndent(*db, "", "\t")
	if err != nil {
		bye(fmt.Sprintf(pad2+"> could not create database file: %s\n", err.Error()), 1)
	}

	path := filepath.Join(workDir, dbFile)

	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		bye(fmt.Sprintf(pad2+"> could not save database file: %s\n", err.Error()), 1)
	}
}

func (db *database) addShowByID(id uint64) error {
	// Query thetvdb.com for show data
	s, err := tvdbFetchShow(id)
	if err != nil {
		return err
	}

	db.Shows[s.ID] = &s
	db.save()

	return nil
}

func (db *database) updateShow(s *Show) error {
	return nil
}

func (db *database) addEpisodesForShow(s *Show) error {
	eps, err := tvdbFetchAllEpisodes(s)
	if err != nil {
		return err
	}

	s.EpisodeIDs = []uint64{}
	for _, ep := range eps {
		db.Episodes[ep.ID] = &ep
		s.EpisodeIDs = append(s.EpisodeIDs, ep.ID)
	}
	db.save()

	return nil
}

type Show struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"seriesName"`
	Status      string   `json:"status"`
	Network     string   `json:"network"`
	FirstAired  string   `json:"firstAired"`
	Runtime     uint32   `json:"runtime,string"`
	Overview    string   `json:"overview"`
	LastUpdated int64    `json:"lastUpdated"`
	LastQuery   int64    `json:"lastQuery"`
	AirDay      string   `json:"airsDayOfWeek"`
	AirTime     string   `json:"airsTime"`
	Rating      float32  `json:"siteRating"`
	RatingCount uint32   `json:"siteRatingCount"`
	EpisodeIDs  []uint64 `json:"episodeIds"`
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
