package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type SliceTracked struct {
	ID    int
	Start *time.Time
	End   *time.Time
}

func getLatestTrack() (*SliceTracked, error) {
	row := db.QueryRow("SELECT id, start, end FROM slice_tracked ORDER BY id DESC LIMIT 1")

	var st SliceTracked
	err := row.Scan(&st.ID, &st.Start, &st.End)
	if err != nil {
		return nil, err
	}

	return &st, nil
}

func updateTrack(track *SliceTracked) error {
	_, err := db.Exec("UPDATE slice_tracked SET end = $1 WHERE id = $2", track.End, track.ID)
	return err
}

func insertTrack(track *SliceTracked) error {
	_, err := db.Exec("INSERT INTO slice_tracked (start) VALUES ($1)", track.Start)
	return err
}

func toggleTracking() error {
	track, err := getLatestTrack()
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			goto start
		}
		return err
	}

	if track.End == nil {
		t := time.Now()
		track.End = &t
		err = updateTrack(track)
		if err != nil {
			return err
		}
		fmt.Println("Tracking stopped")
		return nil
	}

start:
	t := time.Now()
	newTrack := &SliceTracked{
		Start: &t,
	}

	err = insertTrack(newTrack)
	if err != nil {
		return err
	}
	fmt.Println("Tracking started")

	return nil
}

var path string = "C:/work/clockwork/"

func main() {
	var err error
	db, err = sql.Open("sqlite", path+"slice.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	migrations, err := os.ReadFile(path + "migrations.sql")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec(string(migrations))
	if err != nil {
		fmt.Println(err)
	}

	err = toggleTracking()
	if err != nil {
		fmt.Println(err)
	}

	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}
