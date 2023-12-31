package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"database/sql"

	evdev "github.com/gvalkov/golang-evdev"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkgutil/osutil"
)

type EventType = uint64

const (
	EV_KEY EventType = 1
	EV_SYN           = 0
	EV_REL           = 2
	EV_MSC           = 4
)

func GenerateEvents(ctx context.Context, deviceName string) (chan *time.Time, error) {
	device, err := evdev.Open(deviceName)
	if err != nil {
		return nil, fmt.Errorf("opening input device %s", deviceName)
	}

	ch := make(chan *time.Time)
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
				ch <- &t
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
	log.Printf("persisting start=%s end=%s", start, end)
	_, err := d.db.Exec("insert into sessions (start, end) values (?, ?)", start, end)
	if err != nil {
		return fmt.Errorf("inserting session into database: %w", err)
	}
	return nil
}

func (d *SqliteDatabase) Close() {
	d.db.Close()
}

func ensureOutputPath(outputPath string) error {
	dir := path.Dir(outputPath)
	if !osutil.IsDirExist(dir) {
		osutil.MkdirAll(dir, 0777)
	}

	return nil
}

func argError(message string) {
	fmt.Fprintln(os.Stderr, message)
	fmt.Fprintln(os.Stderr, "Usage:")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var (
		inputPathArg = flag.String("flag", "", "Input path")
		dbArg        = flag.String("db", "db.db", "Path to the database")
		timeoutArg   = flag.Float64("timeout", 5, "Seconds to wait before starting new session")
	)
	flag.Parse()

	if *inputPathArg == "" {
		argError("no input path given")
	}

	if err := ensureOutputPath(*dbArg); err != nil {
		log.Fatalf("ensuring output path %s: %v", *dbArg, err)
	}

	ch, err := GenerateEvents(context.TODO(), *inputPathArg)
	if err != nil {
		panic(err)
	}

	db, err := NewDB(*dbArg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("starting collecting readings")
	now := time.Now()
	last := &now
	var sessionStart *time.Time

	for t := range ch {
		// capture first event in session
		if sessionStart == nil {
			log.Printf("starting new session")
			sessionStart = t
		}

		if t.Sub(*last).Seconds() > *timeoutArg {
			log.Printf("got new readings group")
			if err := db.Persist(sessionStart, last); err != nil {
				log.Printf("error persisting to database: %v", err)
			}
			log.Printf("closing session")
			sessionStart = nil
		}
		last = t
	}
}
