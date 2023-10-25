package main

import (
	"log"

	evdev "github.com/gvalkov/golang-evdev"
)

func main() {
	device, err := evdev.Open("/dev/input/by-id/usb-Logitech_G203_LIGHTSYNC_Gaming_Mouse_205935534B58-event-mouse")
	if err != nil {
		panic(err)
	}

	log.Println("printing events")
	for {
		events, err := device.Read()
		if err != nil {
			panic(err)
		}

		for _, event := range events {
			log.Printf("%+v", event)
		}
	}
}
