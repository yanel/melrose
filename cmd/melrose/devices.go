package main

import (
	"log"

	"github.com/emicklei/melrose"
	"github.com/emicklei/melrose/audiolib"
	"github.com/emicklei/melrose/midi"
)

func setupAudio(deviceId string) melrose.AudioDevice {
	if deviceId == "midi" {
		d, err := midi.Open()
		if err != nil {
			log.Fatalln("cannot use audio device:", err)
		}
		return d
	}
	// default
	d, err := audiolib.Open()
	if err != nil {
		log.Fatalln("cannot use audio device:", err)
	}
	d.LoadSounds()
	return d
}
