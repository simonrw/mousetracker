package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"database/sql"
	evdev "github.com/gvalkov/golang-evdev"
	_ "github.com/mattn/go-sqlite3"
)

type EventType = uint64

const (
	EV_KEY EventType = 1
	EV_SYN           = 0
	EV_REL           = 2
	EV_MSC           = 4
)

func GenerateEvents(ctx context.Context, deviceName string) (chan time.Time, error) {
	device, err := evdev.Open(deviceName)
	if err != nil {
		return nil, fmt.Errorf("opening input device %s", deviceName)
	}

	ch := make(chan time.Time)
	go func() {
		for {
			events, err := device.Read()
			if err != nil {
				panic(err)
			}

			for _, event := range events {
				if event.Type != EV_REL {
					continue
				}

				t := time.Unix(event.Time.Sec, event.Time.Usec*1000)
				ch <- t
			}
		}
	}()
	return ch, nil
}

type Database interface {
	Persist(start *time.Time, end *time.Time) error
	Close()
}

type SqliteDatabase struct {
	db *sql.DB
}

func NewDB(path string) (Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}

	d := &SqliteDatabase{
		db,
	}
	if err := d.init(); err != nil {
		return nil, fmt.Errorf("initialising database: %w", err)
	}

	return d, err
}

func (d *SqliteDatabase) init() error {
	_, err := d.db.Exec("create table if not exists sessions (id integer primary key, start datetime not null, end datetime not null)")
	if err != nil {
		return fmt.Errorf("initialising database: %w", err)
	}
	return nil
}

func (d *SqliteDatabase) Persist(start *time.Time, end *time.Time) error {
	_, err := d.db.Exec("insert into sessions (start, end) values (?, ?)", start, end)
	if err != nil {
		return fmt.Errorf("inserting session into database: %w", err)
	}
	return nil
}

func (d *SqliteDatabase) Close() {
	d.db.Close()
}

func main() {
	ch, err := GenerateEvents(context.TODO(), "/dev/input/by-id/usb-Logitech_G203_LIGHTSYNC_Gaming_Mouse_205935534B58-event-mouse")
	if err != nil {
		panic(err)
	}

	db, err := NewDB("db.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var readingsInChunk []time.Time
	last := time.Now()
	log.Printf("starting collecting readings")
	for t := range ch {
		if t.Sub(last).Seconds() > 5 {
			last = time.Now()
			log.Printf("got new readings group with %d entries", len(readingsInChunk))
			if len(readingsInChunk) == 0 {
				// ignore
				continue
			}
			if err := db.Persist(&readingsInChunk[0], &last); err != nil {
				log.Printf("error persisting to database: %v", err)
			}
			readingsInChunk = nil
		}
		readingsInChunk = append(readingsInChunk, t)
	}
}
