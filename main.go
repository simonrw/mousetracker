package main

import (
	"context"
	"fmt"
	"log"
	"time"

	evdev "github.com/gvalkov/golang-evdev"
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
}

type SqliteDatabase struct {
}

func NewDB() Database {
	return &SqliteDatabase{}
}

func (d *SqliteDatabase) Persist(start *time.Time, end *time.Time) error {
	return nil
}

func main() {
	ch, err := GenerateEvents(context.TODO(), "/dev/input/by-id/usb-Logitech_G203_LIGHTSYNC_Gaming_Mouse_205935534B58-event-mouse")
	if err != nil {
		panic(err)
	}

	db := NewDB()
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
