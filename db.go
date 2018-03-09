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
	LastUpdated int64  `json:"lastUpdated"`
	Overview    string `json:"overview"`
}

type database struct {
	Shows       map[uint64]*Show    `json:"shows"`
	Episodes    map[uint64]*episode `json:"episodes"`
	Watched     map[uint64]bool     `json:"watched"`
	LastUpdated int64               `json:"lastUpdated"`
	Version     uint32              `json:"version"`
}

func OpenDB() *database {
	db := &database{
		Shows:       map[uint64]*Show{},
		Episodes:    map[uint64]*episode{},
		Watched:     map[uint64]bool{},
		LastUpdated: time.Now().Unix(),
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
	s, err := tvdbFetchShow(id, false)
	if err != nil {
		return err
	}

	db.Shows[s.ID] = &s
	db.save()

	return nil
}

func (db *database) importFavorites() {
	fmt.Printf("%s> importing all favorite shows..\n", pad2)

	sids, err := tvdbFetchFavorites()
	if err != nil {
		fmt.Printf("failed because of: '%s'\n", err.Error())
		return
	}
	fmt.Println("")

	newShows := 0
	for _, id := range sids {
		if err := db.addShowByID(id); err != nil {
			fmt.Printf("%s> skipping show with ID %d because of: '%s'. ", pad2, id, err.Error())
			continue
		}
		newShows++
	}
	db.save()
	fmt.Printf("%s> added %d shows to database\n\n", pad2, newShows)

	ups := []*Show{}
	for _, s := range db.Shows {
		ups = append(ups, s)
	}
	db.updateShows(ups)
}

func (db *database) updateAllShows() {
	ups := []*Show{}

	fmt.Printf("%s> querying tvdb.com for updates..", pad2)
	// 1. Detect which shows need updating and store them in a slice.
	for _, show := range db.Shows {
		hasUp, err := tvdbHasShowUpdates(show)
		if err != nil {
			// TODO: err could be logged if there was a log. For now we ignore this.
			fmt.Print("x")
		} else {
			if hasUp {
				// The show has updated data on the server.
				// Mark it for updating.
				ups = append(ups, show)
			}
			fmt.Print(".")
		}
	}
	db.save() // Last query time is saved here.
	fmt.Println(" done\n")

	// 2. Update all shows in the ups slice.
	db.updateShows(ups)
}

func (db *database) updateAllEpisodesForShow(s *Show) error {
	eps, err := tvdbFetchAllEpisodes(s)
	if err != nil {
		return err
	}

	s.EpisodeIDs = []uint64{}
	for _, ep := range eps {
		s.EpisodeIDs = append(s.EpisodeIDs, ep.ID)
		db.Episodes[ep.ID] = &ep
	}

	db.save()

	return nil
}

func (db *database) updateShows(ups []*Show) {
	for _, show := range ups {
		fmt.Printf("%s> updating show '%s' (id: %d).. ", pad2, show.Name, show.ID)
		// We just replace the show with the new data,
		// if the request was succesful.
		upShow, err := tvdbFetchShow(show.ID, true)
		if err != nil {
			// TODO: err could be logged if there was a log. For now we ignore this.
			fmt.Println(" failed.")
			continue
		}
		fmt.Println(" done.")

		// If ids on thetvdb.com ever change everything will be bad (they should not).
		db.Shows[show.ID] = &upShow

		// Now we replace the show's episodes.
		err = db.updateAllEpisodesForShow(&upShow)
		if err != nil {
			// TODO: err could be logged if there was a log. For now we ignore this.
			continue
		}
	}
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
